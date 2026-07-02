//from https://github.com/hktkqj/cxdec-hxv4-static-analysis
// Usage: dotnet run --file .\derive.cs -a x86 -- .\bootstrap_dump\Bootstrap.json

using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

return await Derive.RunAsync(args);

internal static class Derive
{
    private const int ManagerSize = 0x30B0;
    private const int TableRva = 0x80E38;
    private const int ArchiveSeedRva = 0x81758;

    private const int RvaManagerCtor = 0x0E2D0;
    private const int RvaHashKeyDerive = 0x10410;
    private const int RvaBootstrapDerive = 0x15630;
    private const int RvaArchiveDerive = 0x157D0;

    private const int ContextOffset = 0x28;
    private const int DripHolderOffset = 0x08;
    private const int DripLaneBaseOffset = 0x04;
    private const int LaneCount = 128;
    private const int LaneSize = 0x10;
    private const int RecordSize = 8;
    private const int CallbackRvaWindow = 0x200000;

    [UnmanagedFunctionPointer(CallingConvention.ThisCall)]
    private delegate nint ManagerCtor(nint thisPtr);

    [UnmanagedFunctionPointer(CallingConvention.ThisCall)]
    private delegate byte BootstrapDerive(
        nint thisPtr,
        nint bootstrapBytes,
        nuint bootstrapSize,
        nint paramsBytes,
        nuint paramsSize);

    [UnmanagedFunctionPointer(CallingConvention.ThisCall)]
    private delegate int ArchiveDerive(
        nint thisPtr,
        nint archiveNameBytes,
        nuint archiveNameSize,
        nint seedBytes);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void HashKeyDerive(
        nint outBytes,
        nuint outSize,
        nint dataBytes,
        nuint dataSize,
        int seed);

    public static async Task<int> RunAsync(string[] argv)
    {
        Console.OutputEncoding = Encoding.UTF8;

        try
        {
            var bootstrapJsonPath = argv[0];
            var summary = await BootstrapSummary.LoadAsync(bootstrapJsonPath);
            var dllPath = Path.ChangeExtension(bootstrapJsonPath, ".dll");
            var outPath = Path.Combine(Path.GetDirectoryName(bootstrapJsonPath)!, "hxv4.json");

            var module = Native.LoadLibraryW(dllPath);
            if (module == 0)
            {
                throw new InvalidOperationException($"LoadLibraryW failed for {dllPath} (Win32={Marshal.GetLastWin32Error()})");
            }

            var config = ReadConfigTable(module + TableRva);
            var warning = DecodeAscii(config["WARNING"]);
            var unique = DecodeUtf16Le(config["UNIQUE"]);
            var bootstrapText = summary.ResolveBootstrapText(warning, unique);
            var archiveText = summary.ResolveArchiveText(unique);
            var archiveSeed = ReadArchiveSeed(module);
            var payload = DerivePayload(
                module,
                dllPath,
                config["PARAMS"],
                Encoding.Unicode.GetBytes(bootstrapText),
                Encoding.Unicode.GetBytes(archiveText));

            Directory.CreateDirectory(Path.GetDirectoryName(outPath)!);
            var json = JsonSerializer.Serialize(payload, DeriveJsonContext.Default.DripProgramPayload);
            await File.WriteAllTextAsync(outPath, json + Environment.NewLine, new UTF8Encoding(false));

            Console.WriteLine($"write hxv4.json success");
            return 0;
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine(ex.Message);
            return 1;
        }
    }

    private static DripProgramPayload DerivePayload(
        nint module,
        string dllPath,
        byte[] paramsBytes,
        byte[] bootstrapBytes,
        byte[] archiveBytes)
    {
        var archiveSeed = ReadArchiveSeed(module);
        var manager = Marshal.AllocHGlobal(ManagerSize);
        try
        {
            ZeroMemory(manager, ManagerSize);
            GetDelegate<ManagerCtor>(module, RvaManagerCtor)(manager);

            var managerCore = manager + DripHolderOffset;
            WithPinned(bootstrapBytes, bootstrapPtr =>
            WithPinned(paramsBytes, paramsPtr =>
            {
                var ok = GetDelegate<BootstrapDerive>(module, RvaBootstrapDerive)(
                    managerCore,
                    bootstrapPtr,
                    (nuint)bootstrapBytes.Length,
                    paramsPtr,
                    (nuint)paramsBytes.Length);
                if (ok == 0)
                {
                    throw new InvalidOperationException("bootstrap derive failed; check Bootstrap.json content");
                }
            }));

            WithPinned(archiveBytes, archivePtr =>
            WithPinned(archiveSeed, seedPtr =>
            {
                GetDelegate<ArchiveDerive>(module, RvaArchiveDerive)(
                    managerCore,
                    archivePtr,
                    (nuint)archiveBytes.Length,
                    seedPtr);
            }));

            return SnapshotPayload(module, dllPath, manager, DeriveHashKey(module, manager));
        }
        finally
        {
            Marshal.FreeHGlobal(manager);
        }
    }

    private static DripProgramPayload SnapshotPayload(nint module, string dllPath, nint manager, byte[] hashKey)
    {
        var dripImpl = ReadU32(manager + DripHolderOffset);
        if (dripImpl == 0)
        {
            throw new InvalidOperationException("DripValueImpl pointer is null after bootstrap derivation");
        }

        var context = manager + ContextOffset;
        var lanes = new List<LanePayload>(LaneCount);

        for (var laneIndex = 0; laneIndex < LaneCount; laneIndex++)
        {
            var lane = (nint)(dripImpl + DripLaneBaseOffset + laneIndex * LaneSize);
            var begin = ReadU32(lane);
            var end = ReadU32(lane + 4);
            if (end < begin || ((end - begin) % RecordSize) != 0)
            {
                throw new InvalidOperationException($"lane {laneIndex} has invalid record range 0x{begin:x8}..0x{end:x8}");
            }

            var records = new List<uint[]>(checked((int)((end - begin) / RecordSize)));
            for (var recordVa = begin; recordVa < end; recordVa += RecordSize)
            {
                var callback = ReadU32((nint)(recordVa + 4));
                if (callback >= unchecked((uint)module) && callback < unchecked((uint)module) + CallbackRvaWindow)
                {
                    callback -= unchecked((uint)module);
                }

                records.Add([
                    ReadU32((nint)recordVa),
                    callback,
                ]);
            }

            lanes.Add(new LanePayload
            {
                Index = laneIndex,
                BeginVa = begin,
                EndVa = end,
                CurrentVa = ReadU32(lane + 8),
                CtxVa = ReadU32(lane + 12),
                Records = records,
            });
        }

        return new DripProgramPayload
        {
            Version = 1,
            SourceModule = Path.GetFileName(dllPath),
            SourceModuleBase = unchecked((uint)module),
            ManagerVa = unchecked((uint)manager),
            DripImplVa = dripImpl,
            Hxv4Key = ToHex(ReadBytes(manager + DripHolderOffset + 0x3038, 32)),
            Hxv4Nonce0 = ToHex(ReadBytes(manager + DripHolderOffset + 0x3078, 24)),
            Hxv4Nonce1 = ToHex(ReadBytes(manager + DripHolderOffset + 0x3058, 24)),
            HashKey = ToHex(hashKey),
            HolderWords = ReadU32List(manager + DripHolderOffset, 6),
            ContextVa = unchecked((uint)context),
            ContextU32 = ReadU32List(context, (ManagerSize - ContextOffset) / 4),
            CallbackRvaBase = unchecked((uint)module),
            Lanes = lanes,
        };
    }

    private static byte[] DeriveHashKey(nint module, nint manager)
    {
        var outBytes = new byte[32];
        if ((ReadU32(manager + 0x30A0) & 3) != 3)
        {
            return outBytes;
        }

        WithPinned(outBytes, outPtr =>
        {
            GetDelegate<HashKeyDerive>(module, RvaHashKeyDerive)(
                outPtr,
                (nuint)outBytes.Length,
                manager + 0x3040,
                0x40,
                -1);
        });
        return outBytes;
    }

    private static byte[] ReadArchiveSeed(nint module)
    {
        var seed = ReadBytes(module + ArchiveSeedRva, 8);
        foreach (var value in seed)
        {
            if (value != 0)
            {
                return seed;
            }
        }
        return ReadFallbackArchiveSeed(module);
    }

    private static byte[] ReadFallbackArchiveSeed(nint module)
    {
        var code = ReadBytes(module + RvaArchiveDerive, 0x80);
        for (var index = 0; index + 14 <= code.Length; index++)
        {
            if (
                code[index] == 0xC7 &&
                code[index + 1] == 0x45 &&
                code[index + 7] == 0xC7 &&
                code[index + 8] == 0x45 &&
                unchecked((byte)(code[index + 2] + 4)) == code[index + 9])
            {
                var seed = new byte[8];
                Array.Copy(code, index + 3, seed, 0, 4);
                Array.Copy(code, index + 10, seed, 4, 4);
                return seed;
            }
        }

        throw new InvalidOperationException("archive seed is zero and fallback constants were not found");
    }

    private static Dictionary<string, byte[]> ReadConfigTable(nint tablePtr)
    {
        var result = new Dictionary<string, byte[]>(StringComparer.Ordinal);
        var cursor = tablePtr;
        while (true)
        {
            var label = ReadAsciiZ(cursor);
            if (label.Length == 0)
            {
                break;
            }

            cursor += label.Length + 1;
            var length = ReadU16(cursor);
            cursor += 2;
            result[label] = ReadBytes(cursor, length);
            cursor += length;
        }

        foreach (var key in new[] { "PARAMS", "UNIQUE", "WARNING" })
        {
            if (!result.ContainsKey(key))
            {
                throw new InvalidOperationException($"config table does not contain {key}");
            }
        }
        return result;
    }

    private static T GetDelegate<T>(nint module, int rva) where T : Delegate
    {
        return Marshal.GetDelegateForFunctionPointer<T>(module + rva);
    }

    private static void WithPinned(byte[] bytes, Action<nint> action)
    {
        var handle = GCHandle.Alloc(bytes, GCHandleType.Pinned);
        try
        {
            action(handle.AddrOfPinnedObject());
        }
        finally
        {
            handle.Free();
        }
    }

    private static void ZeroMemory(nint ptr, int size)
    {
        Marshal.Copy(new byte[size], 0, ptr, size);
    }

    private static ushort ReadU16(nint ptr) => unchecked((ushort)Marshal.ReadInt16(ptr));

    private static uint ReadU32(nint ptr) => unchecked((uint)Marshal.ReadInt32(ptr));

    private static List<uint> ReadU32List(nint ptr, int count)
    {
        var values = new List<uint>(count);
        for (var i = 0; i < count; i++)
        {
            values.Add(ReadU32(ptr + i * 4));
        }
        return values;
    }

    private static byte[] ReadBytes(nint ptr, int size)
    {
        var data = new byte[size];
        Marshal.Copy(ptr, data, 0, size);
        return data;
    }

    private static string ReadAsciiZ(nint ptr)
    {
        var bytes = new List<byte>();
        for (var cursor = ptr; ; cursor++)
        {
            var value = Marshal.ReadByte(cursor);
            if (value == 0)
            {
                return Encoding.ASCII.GetString(bytes.ToArray());
            }
            bytes.Add(value);
        }
    }

    private static string DecodeUtf16Le(byte[] data)
    {
        var end = data.Length;
        while (end >= 2 && data[end - 1] == 0 && data[end - 2] == 0)
        {
            end -= 2;
        }
        return Encoding.Unicode.GetString(data, 0, end);
    }

    private static string DecodeAscii(byte[] data)
    {
        var end = Array.IndexOf(data, (byte)0);
        if (end < 0)
        {
            end = data.Length;
        }
        return Encoding.ASCII.GetString(data, 0, end);
    }

    private static string ToHex(byte[] bytes) => Convert.ToHexString(bytes).ToLowerInvariant();
}

internal sealed class BootstrapSummary
{
    [JsonPropertyName("bootstrap_prefix")]
    public string? BootstrapPrefix { get; set; }

    [JsonPropertyName("warning")]
    public string? Warning { get; set; }

    [JsonPropertyName("final_bootstrap")]
    public string? FinalBootstrap { get; set; }

    [JsonPropertyName("archive_unique_key")]
    public string? ArchiveUniqueKey { get; set; }

    public static async Task<BootstrapSummary> LoadAsync(string path)
    {
        var fullPath = Path.GetFullPath(path);
        var json = await File.ReadAllTextAsync(fullPath, Encoding.UTF8);
        return JsonSerializer.Deserialize(json, DeriveJsonContext.Default.BootstrapSummary)
            ?? throw new InvalidOperationException($"failed to parse {fullPath}");
    }

    public string ResolveBootstrapText(string warning, string unique)
    {
        if (!string.IsNullOrEmpty(FinalBootstrap))
        {
            return FinalBootstrap;
        }
        return (string.IsNullOrEmpty(BootstrapPrefix) ? unique : BootstrapPrefix) + warning;
    }

    public string ResolveArchiveText(string unique)
    {
        return string.IsNullOrEmpty(ArchiveUniqueKey) ? unique : ArchiveUniqueKey;
    }
}

internal sealed class SummaryOutputs
{
    [JsonPropertyName("dll")]
    public string? Dll { get; set; }
}

internal sealed class DripProgramPayload
{
    [JsonPropertyName("version")]
    public int Version { get; set; }

    [JsonPropertyName("source_module")]
    public string SourceModule { get; set; } = "";

    [JsonPropertyName("source_module_base")]
    public uint SourceModuleBase { get; set; }

    [JsonPropertyName("manager_va")]
    public uint ManagerVa { get; set; }

    [JsonPropertyName("drip_impl_va")]
    public uint DripImplVa { get; set; }

    [JsonPropertyName("hxv4_key")]
    public string Hxv4Key { get; set; } = "";

    [JsonPropertyName("hxv4_nonce0")]
    public string Hxv4Nonce0 { get; set; } = "";

    [JsonPropertyName("hxv4_nonce1")]
    public string Hxv4Nonce1 { get; set; } = "";

    [JsonPropertyName("hash_key")]
    public string HashKey { get; set; } = "";

    [JsonPropertyName("holder_words")]
    public List<uint> HolderWords { get; set; } = [];

    [JsonPropertyName("context_va")]
    public uint ContextVa { get; set; }

    [JsonPropertyName("context_u32")]
    public List<uint> ContextU32 { get; set; } = [];

    [JsonPropertyName("callback_rva_base")]
    public uint CallbackRvaBase { get; set; }

    [JsonPropertyName("lanes")]
    public List<LanePayload> Lanes { get; set; } = [];
}

internal sealed class LanePayload
{
    [JsonPropertyName("index")]
    public int Index { get; set; }

    [JsonPropertyName("begin_va")]
    public uint BeginVa { get; set; }

    [JsonPropertyName("end_va")]
    public uint EndVa { get; set; }

    [JsonPropertyName("current_va")]
    public uint CurrentVa { get; set; }

    [JsonPropertyName("ctx_va")]
    public uint CtxVa { get; set; }

    [JsonPropertyName("records")]
    public List<uint[]> Records { get; set; } = [];
}

internal static class Native
{
    [DllImport("kernel32", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern nint LoadLibraryW(string path);
}

[JsonSourceGenerationOptions(WriteIndented = true)]
[JsonSerializable(typeof(BootstrapSummary))]
[JsonSerializable(typeof(SummaryOutputs))]
[JsonSerializable(typeof(DripProgramPayload))]
[JsonSerializable(typeof(LanePayload))]
internal partial class DeriveJsonContext : JsonSerializerContext
{
}
