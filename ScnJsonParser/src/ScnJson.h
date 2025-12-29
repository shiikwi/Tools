//  To parse this JSON data, first install
//
//      json.hpp  https://github.com/nlohmann/json
//
//  Then include this file, and then do
//
//     ScnJson data = nlohmann::json::parse(jsonString);

#pragma once

#include <optional>
#include <variant>
#include <nlohmann/json.hpp>

#include <optional>
#include <stdexcept>
#include <regex>

#ifndef NLOHMANN_OPT_HELPER
#define NLOHMANN_OPT_HELPER
namespace nlohmann {
    template <typename T>
    struct adl_serializer<std::shared_ptr<T>> {
        static void to_json(json & j, std::shared_ptr<T> const & opt) {
            if (!opt) j = nullptr; else j = *opt;
        }

        static std::shared_ptr<T> from_json(json const & j) {
            if (j.is_null()) return std::make_shared<T>(); else return std::make_shared<T>(j.get<T>());
        }
    };
    template <typename T>
    struct adl_serializer<std::optional<T>> {
        static void to_json(json & j, std::optional<T> const & opt) {
            if (!opt) j = nullptr; else j = *opt;
        }

        static std::optional<T> from_json(json const & j) {
            if (j.is_null()) return std::make_optional<T>(); else return std::make_optional<T>(j.get<T>());
        }
    };
}
#endif

namespace Scn {
    using nlohmann::json;

    #ifndef NLOHMANN_UNTYPED_Scn_HELPER
    #define NLOHMANN_UNTYPED_Scn_HELPER
    inline json get_untyped(json const & j, char const * property) {
        if (j.find(property) != j.end()) {
            return j.at(property).get<json>();
        }
        return json();
    }

    inline json get_untyped(json const & j, std::string property) {
        return get_untyped(j, property.data());
    }
    #endif

    #ifndef NLOHMANN_OPTIONAL_Scn_HELPER
    #define NLOHMANN_OPTIONAL_Scn_HELPER
    template <typename T>
    inline std::shared_ptr<T> get_heap_optional(json const & j, char const * property) {
        auto it = j.find(property);
        if (it != j.end() && !it->is_null()) {
            return j.at(property).get<std::shared_ptr<T>>();
        }
        return std::shared_ptr<T>();
    }

    template <typename T>
    inline std::shared_ptr<T> get_heap_optional(json const & j, std::string property) {
        return get_heap_optional<T>(j, property.data());
    }
    template <typename T>
    inline std::optional<T> get_stack_optional(json const & j, char const * property) {
        auto it = j.find(property);
        if (it != j.end() && !it->is_null()) {
            return j.at(property).get<std::optional<T>>();
        }
        return std::optional<T>();
    }

    template <typename T>
    inline std::optional<T> get_stack_optional(json const & j, std::string property) {
        return get_stack_optional<T>(j, property.data());
    }
    #endif

    class Llmap {
        public:
        Llmap() = default;
        virtual ~Llmap() = default;

        private:
        std::optional<std::vector<int64_t>> empty;
        std::optional<std::vector<int64_t>> llmap;
        std::optional<std::vector<int64_t>> purple;
        std::optional<std::vector<int64_t>> fluffy;
        std::optional<std::vector<int64_t>> tentacled;
        std::optional<std::vector<int64_t>> sticky;
        std::optional<std::vector<int64_t>> indigo;
        std::optional<std::vector<int64_t>> indecent;
        std::optional<std::vector<int64_t>> hilarious;
        std::optional<std::vector<int64_t>> ambitious;
        std::optional<std::string> name;

        public:
        std::optional<std::vector<int64_t>> get_empty() const { return empty; }
        void set_empty(std::optional<std::vector<int64_t>> value) { this->empty = value; }

        std::optional<std::vector<int64_t>> get_llmap() const { return llmap; }
        void set_llmap(std::optional<std::vector<int64_t>> value) { this->llmap = value; }

        std::optional<std::vector<int64_t>> get_purple() const { return purple; }
        void set_purple(std::optional<std::vector<int64_t>> value) { this->purple = value; }

        std::optional<std::vector<int64_t>> get_fluffy() const { return fluffy; }
        void set_fluffy(std::optional<std::vector<int64_t>> value) { this->fluffy = value; }

        std::optional<std::vector<int64_t>> get_tentacled() const { return tentacled; }
        void set_tentacled(std::optional<std::vector<int64_t>> value) { this->tentacled = value; }

        std::optional<std::vector<int64_t>> get_sticky() const { return sticky; }
        void set_sticky(std::optional<std::vector<int64_t>> value) { this->sticky = value; }

        std::optional<std::vector<int64_t>> get_indigo() const { return indigo; }
        void set_indigo(std::optional<std::vector<int64_t>> value) { this->indigo = value; }

        std::optional<std::vector<int64_t>> get_indecent() const { return indecent; }
        void set_indecent(std::optional<std::vector<int64_t>> value) { this->indecent = value; }

        std::optional<std::vector<int64_t>> get_hilarious() const { return hilarious; }
        void set_hilarious(std::optional<std::vector<int64_t>> value) { this->hilarious = value; }

        std::optional<std::vector<int64_t>> get_ambitious() const { return ambitious; }
        void set_ambitious(std::optional<std::vector<int64_t>> value) { this->ambitious = value; }

        std::optional<std::string> get_name() const { return name; }
        void set_name(std::optional<std::string> value) { this->name = value; }
    };

    enum class StorageEnum : int { BSIS_EV_001_TXT, GAMELOOP_TXT };

    enum class ClassEnum : int { BG, BGM, CHARACTER, DEL, EMPTY, ENV, INIT, LOOPSE, LSE1, NEW, STAGE };

    using IndigoLine = std::variant<ClassEnum, int64_t>;

    using Time = std::variant<int64_t, std::string>;

    enum class Handler : int { LOOP, MOVE_ACTION, SIN_ACTION, WAIT };

    class ActionClass {
        public:
        ActionClass() = default;
        virtual ~ActionClass() = default;

        private:
        std::optional<int64_t> delay;
        std::optional<Handler> handler;
        std::optional<int64_t> start;
        std::optional<int64_t> time;
        std::optional<Time> value;
        std::optional<Time> accel;
        std::optional<std::string> cycle;
        std::optional<std::string> nofast;
        std::optional<std::string> nowait;
        std::optional<std::string> vibration;
        std::optional<int64_t> count;
        std::optional<int64_t> loop;

        public:
        std::optional<int64_t> get_delay() const { return delay; }
        void set_delay(std::optional<int64_t> value) { this->delay = value; }

        std::optional<Handler> get_handler() const { return handler; }
        void set_handler(std::optional<Handler> value) { this->handler = value; }

        std::optional<int64_t> get_start() const { return start; }
        void set_start(std::optional<int64_t> value) { this->start = value; }

        std::optional<int64_t> get_time() const { return time; }
        void set_time(std::optional<int64_t> value) { this->time = value; }

        std::optional<Time> get_value() const { return value; }
        void set_value(std::optional<Time> value) { this->value = value; }

        std::optional<Time> get_accel() const { return accel; }
        void set_accel(std::optional<Time> value) { this->accel = value; }

        std::optional<std::string> get_cycle() const { return cycle; }
        void set_cycle(std::optional<std::string> value) { this->cycle = value; }

        std::optional<std::string> get_nofast() const { return nofast; }
        void set_nofast(std::optional<std::string> value) { this->nofast = value; }

        std::optional<std::string> get_nowait() const { return nowait; }
        void set_nowait(std::optional<std::string> value) { this->nowait = value; }

        std::optional<std::string> get_vibration() const { return vibration; }
        void set_vibration(std::optional<std::string> value) { this->vibration = value; }

        std::optional<int64_t> get_count() const { return count; }
        void set_count(std::optional<int64_t> value) { this->count = value; }

        std::optional<int64_t> get_loop() const { return loop; }
        void set_loop(std::optional<int64_t> value) { this->loop = value; }
    };

    enum class ActionEnum : int { EMPTY, FADE, LEVELOFFSET, THE_03701058320104000000000000000, THE_03801057920103200000000000000, VISVALUE, XPOS, YPOS, ZOOMX, ZOOMY, ZPOS };

    using LineAction = std::variant<std::vector<ActionClass>, double, ActionEnum>;

    enum class File : int { BG02, L_BLACK, STAND };

    enum class Dress : int { EMPTY };

    enum class Pose : int { EMPTY, POSE, PURPLE };

    class Options {
        public:
        Options() = default;
        virtual ~Options() = default;

        private:
        std::optional<Dress> dress;
        std::optional<std::string> face;
        std::optional<Pose> pose;

        public:
        std::optional<Dress> get_dress() const { return dress; }
        void set_dress(std::optional<Dress> value) { this->dress = value; }

        std::optional<std::string> get_face() const { return face; }
        void set_face(std::optional<std::string> value) { this->face = value; }

        std::optional<Pose> get_pose() const { return pose; }
        void set_pose(std::optional<Pose> value) { this->pose = value; }
    };

    class ImageFile {
        public:
        ImageFile() = default;
        virtual ~ImageFile() = default;

        private:
        std::optional<File> file;
        std::optional<Options> options;

        public:
        std::optional<File> get_file() const { return file; }
        void set_file(std::optional<File> value) { this->file = value; }

        std::optional<Options> get_options() const { return options; }
        void set_options(std::optional<Options> value) { this->options = value; }
    };

    class Redraw {
        public:
        Redraw() = default;
        virtual ~Redraw() = default;

        private:
        std::optional<int64_t> disp;
        std::optional<ImageFile> image_file;
        std::optional<std::string> pos_name;

        public:
        std::optional<int64_t> get_disp() const { return disp; }
        void set_disp(std::optional<int64_t> value) { this->disp = value; }

        std::optional<ImageFile> get_image_file() const { return image_file; }
        void set_image_file(std::optional<ImageFile> value) { this->image_file = value; }

        std::optional<std::string> get_pos_name() const { return pos_name; }
        void set_pos_name(std::optional<std::string> value) { this->pos_name = value; }
    };

    enum class Filename : int { BGM10, EMPTY };

    class Replay {
        public:
        Replay() = default;
        virtual ~Replay() = default;

        private:
        std::optional<Filename> filename;
        std::optional<int64_t> loop;
        nlohmann::json start;
        std::optional<int64_t> state;
        std::optional<int64_t> volume;

        public:
        std::optional<Filename> get_filename() const { return filename; }
        void set_filename(std::optional<Filename> value) { this->filename = value; }

        std::optional<int64_t> get_loop() const { return loop; }
        void set_loop(std::optional<int64_t> value) { this->loop = value; }

        nlohmann::json const & get_start() const { return start; }
        nlohmann::json & get_mutable_start() { return start; }
        void set_start(nlohmann::json const & value) { this->start = value; }

        std::optional<int64_t> get_state() const { return state; }
        void set_state(std::optional<int64_t> value) { this->state = value; }

        std::optional<int64_t> get_volume() const { return volume; }
        void set_volume(std::optional<int64_t> value) { this->volume = value; }
    };

    enum class Method : int { CROSSFADE, EMPTY };

    class Trans {
        public:
        Trans() = default;
        virtual ~Trans() = default;

        private:
        std::optional<Method> method;
        std::optional<Time> time;
        std::optional<int64_t> sync;

        public:
        std::optional<Method> get_method() const { return method; }
        void set_method(std::optional<Method> value) { this->method = value; }

        std::optional<Time> get_time() const { return time; }
        void set_time(std::optional<Time> value) { this->time = value; }

        std::optional<int64_t> get_sync() const { return sync; }
        void set_sync(std::optional<int64_t> value) { this->sync = value; }
    };

    class Update {
        public:
        Update() = default;
        virtual ~Update() = default;

        private:
        std::optional<int64_t> state;

        public:
        std::optional<int64_t> get_state() const { return state; }
        void set_state(std::optional<int64_t> value) { this->state = value; }
    };

    class PurpleLine {
        public:
        PurpleLine() = default;
        virtual ~PurpleLine() = default;

        private:
        std::optional<std::vector<std::optional<std::vector<LineAction>>>> action;
        std::optional<ClassEnum> name;
        std::optional<ClassEnum> line_class;
        std::optional<std::string> link;
        std::optional<Redraw> redraw;
        std::optional<int64_t> showmode;
        nlohmann::json type;
        std::optional<Replay> replay;
        std::optional<Trans> trans;
        std::optional<Update> update;

        public:
        std::optional<std::vector<std::optional<std::vector<LineAction>>>> get_action() const { return action; }
        void set_action(std::optional<std::vector<std::optional<std::vector<LineAction>>>> value) { this->action = value; }

        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }

        std::optional<ClassEnum> get_line_class() const { return line_class; }
        void set_line_class(std::optional<ClassEnum> value) { this->line_class = value; }

        std::optional<std::string> get_link() const { return link; }
        void set_link(std::optional<std::string> value) { this->link = value; }

        std::optional<Redraw> get_redraw() const { return redraw; }
        void set_redraw(std::optional<Redraw> value) { this->redraw = value; }

        std::optional<int64_t> get_showmode() const { return showmode; }
        void set_showmode(std::optional<int64_t> value) { this->showmode = value; }

        nlohmann::json const & get_type() const { return type; }
        nlohmann::json & get_mutable_type() { return type; }
        void set_type(nlohmann::json const & value) { this->type = value; }

        std::optional<Replay> get_replay() const { return replay; }
        void set_replay(std::optional<Replay> value) { this->replay = value; }

        std::optional<Trans> get_trans() const { return trans; }
        void set_trans(std::optional<Trans> value) { this->trans = value; }

        std::optional<Update> get_update() const { return update; }
        void set_update(std::optional<Update> value) { this->update = value; }
    };

    using StickyLine = std::variant<std::vector<IndigoLine>, PurpleLine>;

    using DatumAction = std::variant<double, ActionEnum>;

    class DatumClass {
        public:
        DatumClass() = default;
        virtual ~DatumClass() = default;

        private:
        std::optional<ClassEnum> name;
        std::optional<Replay> replay;
        std::optional<Update> update;
        std::optional<std::vector<std::vector<DatumAction>>> action;
        std::optional<ClassEnum> datum_class;
        std::optional<Redraw> redraw;
        std::optional<int64_t> showmode;
        nlohmann::json type;

        public:
        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }

        std::optional<Replay> get_replay() const { return replay; }
        void set_replay(std::optional<Replay> value) { this->replay = value; }

        std::optional<Update> get_update() const { return update; }
        void set_update(std::optional<Update> value) { this->update = value; }

        std::optional<std::vector<std::vector<DatumAction>>> get_action() const { return action; }
        void set_action(std::optional<std::vector<std::vector<DatumAction>>> value) { this->action = value; }

        std::optional<ClassEnum> get_datum_class() const { return datum_class; }
        void set_datum_class(std::optional<ClassEnum> value) { this->datum_class = value; }

        std::optional<Redraw> get_redraw() const { return redraw; }
        void set_redraw(std::optional<Redraw> value) { this->redraw = value; }

        std::optional<int64_t> get_showmode() const { return showmode; }
        void set_showmode(std::optional<int64_t> value) { this->showmode = value; }

        nlohmann::json const & get_type() const { return type; }
        nlohmann::json & get_mutable_type() { return type; }
        void set_type(nlohmann::json const & value) { this->type = value; }
    };

    using DatumElement = std::variant<DatumClass, ClassEnum>;

    class LineEnv {
        public:
        LineEnv() = default;
        virtual ~LineEnv() = default;

        private:
        std::optional<std::vector<nlohmann::json>> action;
        std::optional<ClassEnum> name;

        public:
        std::optional<std::vector<nlohmann::json>> get_action() const { return action; }
        void set_action(std::optional<std::vector<nlohmann::json>> value) { this->action = value; }

        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }
    };

    class List {
        public:
        List() = default;
        virtual ~List() = default;

        private:
        std::optional<int64_t> mode;
        std::optional<ClassEnum> name;

        public:
        std::optional<int64_t> get_mode() const { return mode; }
        void set_mode(std::optional<int64_t> value) { this->mode = value; }

        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }
    };

    class FluffyLine {
        public:
        FluffyLine() = default;
        virtual ~FluffyLine() = default;

        private:
        std::optional<std::vector<std::vector<DatumElement>>> data;
        std::optional<LineEnv> env;
        std::optional<std::vector<std::optional<List>>> list;
        std::optional<int64_t> wait;
        std::optional<Method> method;
        std::optional<Time> time;

        public:
        std::optional<std::vector<std::vector<DatumElement>>> get_data() const { return data; }
        void set_data(std::optional<std::vector<std::vector<DatumElement>>> value) { this->data = value; }

        std::optional<LineEnv> get_env() const { return env; }
        void set_env(std::optional<LineEnv> value) { this->env = value; }

        std::optional<std::vector<std::optional<List>>> get_list() const { return list; }
        void set_list(std::optional<std::vector<std::optional<List>>> value) { this->list = value; }

        std::optional<int64_t> get_wait() const { return wait; }
        void set_wait(std::optional<int64_t> value) { this->wait = value; }

        std::optional<Method> get_method() const { return method; }
        void set_method(std::optional<Method> value) { this->method = value; }

        std::optional<Time> get_time() const { return time; }
        void set_time(std::optional<Time> value) { this->time = value; }
    };

    using TentacledLine = std::optional<std::variant<std::vector<StickyLine>, FluffyLine, int64_t, std::string>>;

    enum class LineEnum : int { BGM_BGM10, CG_BG02, CG_STAND };

    using SceneLine = std::variant<std::vector<TentacledLine>, LineEnum, int64_t>;

    enum class Eval : int { EVENT_MANAGER_HAS_SEQUENCE_SCENARIO, F_GAMELOOP_0 };

    enum class Evalstorage : int { GET_NEXT_SEQUENCE, S_SPRINTF_HAN2_ZEN_STRING_INTRANDOM_110 };

    class Next {
        public:
        Next() = default;
        virtual ~Next() = default;

        private:
        std::optional<Evalstorage> evalstorage;
        std::optional<int64_t> type;
        std::optional<StorageEnum> storage;
        std::optional<std::string> target;
        std::optional<Eval> eval;

        public:
        std::optional<Evalstorage> get_evalstorage() const { return evalstorage; }
        void set_evalstorage(std::optional<Evalstorage> value) { this->evalstorage = value; }

        std::optional<int64_t> get_type() const { return type; }
        void set_type(std::optional<int64_t> value) { this->type = value; }

        std::optional<StorageEnum> get_storage() const { return storage; }
        void set_storage(std::optional<StorageEnum> value) { this->storage = value; }

        std::optional<std::string> get_target() const { return target; }
        void set_target(std::optional<std::string> value) { this->target = value; }

        std::optional<Eval> get_eval() const { return eval; }
        void set_eval(std::optional<Eval> value) { this->eval = value; }
    };

    using StickyDialogueLine = std::optional<std::variant<int64_t, std::string>>;
    using StickyText = std::vector<StickyDialogueLine>;

    class PurpleText {
        public:
        PurpleText() = default;
        virtual ~PurpleText() = default;

        private:
        std::optional<ClassEnum> name;
        std::optional<int64_t> pan;
        std::optional<int64_t> type;
        std::optional<std::string> voice;

        public:
        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }

        std::optional<int64_t> get_pan() const { return pan; }
        void set_pan(std::optional<int64_t> value) { this->pan = value; }

        std::optional<int64_t> get_type() const { return type; }
        void set_type(std::optional<int64_t> value) { this->type = value; }

        std::optional<std::string> get_voice() const { return voice; }
        void set_voice(std::optional<std::string> value) { this->voice = value; }
    };

    using TentacledText = std::variant<std::string, std::vector<StickyText>, PurpleText>;

    class TextEnv {
        public:
        TextEnv() = default;
        virtual ~TextEnv() = default;

        private:
        std::optional<ClassEnum> name;

        public:
        std::optional<ClassEnum> get_name() const { return name; }
        void set_name(std::optional<ClassEnum> value) { this->name = value; }
    };

    class FluffyText {
        public:
        FluffyText() = default;
        virtual ~FluffyText() = default;

        private:
        std::optional<std::vector<std::vector<DatumElement>>> data;
        std::optional<TextEnv> env;

        public:
        std::optional<std::vector<std::vector<DatumElement>>> get_data() const { return data; }
        void set_data(std::optional<std::vector<std::vector<DatumElement>>> value) { this->data = value; }

        std::optional<TextEnv> get_env() const { return env; }
        void set_env(std::optional<TextEnv> value) { this->env = value; }
    };

    enum class TextEnum : int { EMPTY, TEXT };

    using SceneText = std::optional<std::variant<std::vector<TentacledText>, FluffyText, TextEnum, int64_t>>;

    enum class Title : int { EMPTY, REST, TITLE };

    class Scene {
        public:
        Scene() = default;
        virtual ~Scene() = default;

        private:
        std::optional<int64_t> first_line;
        std::optional<std::string> label;
        std::optional<std::vector<SceneLine>> lines;
        std::optional<std::vector<Next>> nexts;
        std::optional<int64_t> sp_count;
        std::optional<std::vector<Title>> title;
        std::optional<int64_t> version;
        std::optional<std::vector<std::vector<SceneText>>> texts;

        public:
        std::optional<int64_t> get_first_line() const { return first_line; }
        void set_first_line(std::optional<int64_t> value) { this->first_line = value; }

        std::optional<std::string> get_label() const { return label; }
        void set_label(std::optional<std::string> value) { this->label = value; }

        std::optional<std::vector<SceneLine>> get_lines() const { return lines; }
        void set_lines(std::optional<std::vector<SceneLine>> value) { this->lines = value; }

        std::optional<std::vector<Next>> get_nexts() const { return nexts; }
        void set_nexts(std::optional<std::vector<Next>> value) { this->nexts = value; }

        std::optional<int64_t> get_sp_count() const { return sp_count; }
        void set_sp_count(std::optional<int64_t> value) { this->sp_count = value; }

        std::optional<std::vector<Title>> get_title() const { return title; }
        void set_title(std::optional<std::vector<Title>> value) { this->title = value; }

        std::optional<int64_t> get_version() const { return version; }
        void set_version(std::optional<int64_t> value) { this->version = value; }

        std::optional<std::vector<std::vector<SceneText>>> get_texts() const { return texts; }
        void set_texts(std::optional<std::vector<std::vector<SceneText>>> value) { this->texts = value; }
    };

    class ScnJson {
        public:
        ScnJson() = default;
        virtual ~ScnJson() = default;

        private:
        std::optional<std::string> hash;
        std::optional<std::vector<std::string>> languages;
        std::optional<std::vector<Llmap>> llmap;
        std::optional<std::string> name;
        std::optional<std::vector<nlohmann::json>> outlines;
        std::optional<std::vector<Scene>> scenes;

        public:
        std::optional<std::string> get_hash() const { return hash; }
        void set_hash(std::optional<std::string> value) { this->hash = value; }

        std::optional<std::vector<std::string>> get_languages() const { return languages; }
        void set_languages(std::optional<std::vector<std::string>> value) { this->languages = value; }

        std::optional<std::vector<Llmap>> get_llmap() const { return llmap; }
        void set_llmap(std::optional<std::vector<Llmap>> value) { this->llmap = value; }

        std::optional<std::string> get_name() const { return name; }
        void set_name(std::optional<std::string> value) { this->name = value; }

        std::optional<std::vector<nlohmann::json>> get_outlines() const { return outlines; }
        void set_outlines(std::optional<std::vector<nlohmann::json>> value) { this->outlines = value; }

        std::optional<std::vector<Scene>> get_scenes() const { return scenes; }
        void set_scenes(std::optional<std::vector<Scene>> value) { this->scenes = value; }
    };
}

namespace Scn {
void from_json(json const & j, Llmap & x);
void to_json(json & j, Llmap const & x);

void from_json(json const & j, ActionClass & x);
void to_json(json & j, ActionClass const & x);

void from_json(json const & j, Options & x);
void to_json(json & j, Options const & x);

void from_json(json const & j, ImageFile & x);
void to_json(json & j, ImageFile const & x);

void from_json(json const & j, Redraw & x);
void to_json(json & j, Redraw const & x);

void from_json(json const & j, Replay & x);
void to_json(json & j, Replay const & x);

void from_json(json const & j, Trans & x);
void to_json(json & j, Trans const & x);

void from_json(json const & j, Update & x);
void to_json(json & j, Update const & x);

void from_json(json const & j, PurpleLine & x);
void to_json(json & j, PurpleLine const & x);

void from_json(json const & j, DatumClass & x);
void to_json(json & j, DatumClass const & x);

void from_json(json const & j, LineEnv & x);
void to_json(json & j, LineEnv const & x);

void from_json(json const & j, List & x);
void to_json(json & j, List const & x);

void from_json(json const & j, FluffyLine & x);
void to_json(json & j, FluffyLine const & x);

void from_json(json const & j, Next & x);
void to_json(json & j, Next const & x);

void from_json(json const & j, PurpleText & x);
void to_json(json & j, PurpleText const & x);

void from_json(json const & j, TextEnv & x);
void to_json(json & j, TextEnv const & x);

void from_json(json const & j, FluffyText & x);
void to_json(json & j, FluffyText const & x);

void from_json(json const & j, Scene & x);
void to_json(json & j, Scene const & x);

void from_json(json const & j, ScnJson & x);
void to_json(json & j, ScnJson const & x);

void from_json(json const & j, StorageEnum & x);
void to_json(json & j, StorageEnum const & x);

void from_json(json const & j, ClassEnum & x);
void to_json(json & j, ClassEnum const & x);

void from_json(json const & j, Handler & x);
void to_json(json & j, Handler const & x);

void from_json(json const & j, ActionEnum & x);
void to_json(json & j, ActionEnum const & x);

void from_json(json const & j, File & x);
void to_json(json & j, File const & x);

void from_json(json const & j, Dress & x);
void to_json(json & j, Dress const & x);

void from_json(json const & j, Pose & x);
void to_json(json & j, Pose const & x);

void from_json(json const & j, Filename & x);
void to_json(json & j, Filename const & x);

void from_json(json const & j, Method & x);
void to_json(json & j, Method const & x);

void from_json(json const & j, LineEnum & x);
void to_json(json & j, LineEnum const & x);

void from_json(json const & j, Eval & x);
void to_json(json & j, Eval const & x);

void from_json(json const & j, Evalstorage & x);
void to_json(json & j, Evalstorage const & x);

void from_json(json const & j, TextEnum & x);
void to_json(json & j, TextEnum const & x);

void from_json(json const & j, Title & x);
void to_json(json & j, Title const & x);
}
namespace nlohmann {
template <>
struct adl_serializer<std::variant<Scn::ClassEnum, int64_t>> {
    static void from_json(json const & j, std::variant<Scn::ClassEnum, int64_t> & x);
    static void to_json(json & j, std::variant<Scn::ClassEnum, int64_t> const & x);
};

template <>
struct adl_serializer<std::variant<int64_t, std::string>> {
    static void from_json(json const & j, std::variant<int64_t, std::string> & x);
    static void to_json(json & j, std::variant<int64_t, std::string> const & x);
};

template <>
struct adl_serializer<std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum> const & x);
};

template <>
struct adl_serializer<std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine> const & x);
};

template <>
struct adl_serializer<std::variant<double, Scn::ActionEnum>> {
    static void from_json(json const & j, std::variant<double, Scn::ActionEnum> & x);
    static void to_json(json & j, std::variant<double, Scn::ActionEnum> const & x);
};

template <>
struct adl_serializer<std::variant<Scn::DatumClass, Scn::ClassEnum>> {
    static void from_json(json const & j, std::variant<Scn::DatumClass, Scn::ClassEnum> & x);
    static void to_json(json & j, std::variant<Scn::DatumClass, Scn::ClassEnum> const & x);
};

template <>
struct adl_serializer<std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string> const & x);
};

template <>
struct adl_serializer<std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t> const & x);
};

template <>
struct adl_serializer<std::variant<std::vector<Scn::StickyText>, Scn::PurpleText>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::StickyText>, Scn::PurpleText> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::StickyText>, Scn::PurpleText> const & x);
};

template <>
struct adl_serializer<std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText>> {
    static void from_json(json const & j, std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText> & x);
    static void to_json(json & j, std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText> const & x);
};


template <>
struct adl_serializer<std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t>> {
    static void from_json(json const & j, std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t> & x);
    static void to_json(json & j, std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t> const & x);
};
}
namespace Scn {
    inline void from_json(json const & j, Llmap& x) {
        x.set_empty(get_stack_optional<std::vector<int64_t>>(j, "*□１"));
        x.set_llmap(get_stack_optional<std::vector<int64_t>>(j, "*□１０"));
        x.set_purple(get_stack_optional<std::vector<int64_t>>(j, "*□２"));
        x.set_fluffy(get_stack_optional<std::vector<int64_t>>(j, "*□３"));
        x.set_tentacled(get_stack_optional<std::vector<int64_t>>(j, "*□４"));
        x.set_sticky(get_stack_optional<std::vector<int64_t>>(j, "*□５"));
        x.set_indigo(get_stack_optional<std::vector<int64_t>>(j, "*□６"));
        x.set_indecent(get_stack_optional<std::vector<int64_t>>(j, "*□７"));
        x.set_hilarious(get_stack_optional<std::vector<int64_t>>(j, "*□８"));
        x.set_ambitious(get_stack_optional<std::vector<int64_t>>(j, "*□９"));
        x.set_name(get_stack_optional<std::string>(j, "name"));
    }

    inline void to_json(json & j, Llmap const & x) {
        j = json::object();
        j["*□１"] = x.get_empty();
        j["*□１０"] = x.get_llmap();
        j["*□２"] = x.get_purple();
        j["*□３"] = x.get_fluffy();
        j["*□４"] = x.get_tentacled();
        j["*□５"] = x.get_sticky();
        j["*□６"] = x.get_indigo();
        j["*□７"] = x.get_indecent();
        j["*□８"] = x.get_hilarious();
        j["*□９"] = x.get_ambitious();
        j["name"] = x.get_name();
    }

    inline void from_json(json const & j, ActionClass& x) {
        x.set_delay(get_stack_optional<int64_t>(j, "delay"));
        x.set_handler(get_stack_optional<Handler>(j, "handler"));
        x.set_start(get_stack_optional<int64_t>(j, "start"));
        x.set_time(get_stack_optional<int64_t>(j, "time"));
        x.set_value(get_stack_optional<std::variant<int64_t, std::string>>(j, "value"));
        x.set_accel(get_stack_optional<std::variant<int64_t, std::string>>(j, "accel"));
        x.set_cycle(get_stack_optional<std::string>(j, "cycle"));
        x.set_nofast(get_stack_optional<std::string>(j, "nofast"));
        x.set_nowait(get_stack_optional<std::string>(j, "nowait"));
        x.set_vibration(get_stack_optional<std::string>(j, "vibration"));
        x.set_count(get_stack_optional<int64_t>(j, "count"));
        x.set_loop(get_stack_optional<int64_t>(j, "loop"));
    }

    inline void to_json(json & j, ActionClass const & x) {
        j = json::object();
        j["delay"] = x.get_delay();
        j["handler"] = x.get_handler();
        j["start"] = x.get_start();
        j["time"] = x.get_time();
        j["value"] = x.get_value();
        j["accel"] = x.get_accel();
        j["cycle"] = x.get_cycle();
        j["nofast"] = x.get_nofast();
        j["nowait"] = x.get_nowait();
        j["vibration"] = x.get_vibration();
        j["count"] = x.get_count();
        j["loop"] = x.get_loop();
    }

    inline void from_json(json const & j, Options& x) {
        x.set_dress(get_stack_optional<Dress>(j, "dress"));
        x.set_face(get_stack_optional<std::string>(j, "face"));
        x.set_pose(get_stack_optional<Pose>(j, "pose"));
    }

    inline void to_json(json & j, Options const & x) {
        j = json::object();
        j["dress"] = x.get_dress();
        j["face"] = x.get_face();
        j["pose"] = x.get_pose();
    }

    inline void from_json(json const & j, ImageFile& x) {
        x.set_file(get_stack_optional<File>(j, "file"));
        x.set_options(get_stack_optional<Options>(j, "options"));
    }

    inline void to_json(json & j, ImageFile const & x) {
        j = json::object();
        j["file"] = x.get_file();
        j["options"] = x.get_options();
    }

    inline void from_json(json const & j, Redraw& x) {
        x.set_disp(get_stack_optional<int64_t>(j, "disp"));
        x.set_image_file(get_stack_optional<ImageFile>(j, "imageFile"));
        x.set_pos_name(get_stack_optional<std::string>(j, "posName"));
    }

    inline void to_json(json & j, Redraw const & x) {
        j = json::object();
        j["disp"] = x.get_disp();
        j["imageFile"] = x.get_image_file();
        j["posName"] = x.get_pos_name();
    }

    inline void from_json(json const & j, Replay& x) {
        x.set_filename(get_stack_optional<Filename>(j, "filename"));
        x.set_loop(get_stack_optional<int64_t>(j, "loop"));
        x.set_start(get_untyped(j, "start"));
        x.set_state(get_stack_optional<int64_t>(j, "state"));
        x.set_volume(get_stack_optional<int64_t>(j, "volume"));
    }

    inline void to_json(json & j, Replay const & x) {
        j = json::object();
        j["filename"] = x.get_filename();
        j["loop"] = x.get_loop();
        j["start"] = x.get_start();
        j["state"] = x.get_state();
        j["volume"] = x.get_volume();
    }

    inline void from_json(json const & j, Trans& x) {
        x.set_method(get_stack_optional<Method>(j, "method"));
        x.set_time(get_stack_optional<std::variant<int64_t, std::string>>(j, "time"));
        x.set_sync(get_stack_optional<int64_t>(j, "sync"));
    }

    inline void to_json(json & j, Trans const & x) {
        j = json::object();
        j["method"] = x.get_method();
        j["time"] = x.get_time();
        j["sync"] = x.get_sync();
    }

    inline void from_json(json const & j, Update& x) {
        x.set_state(get_stack_optional<int64_t>(j, "state"));
    }

    inline void to_json(json & j, Update const & x) {
        j = json::object();
        j["state"] = x.get_state();
    }

    inline void from_json(json const & j, PurpleLine& x) {
        x.set_action(get_stack_optional<std::vector<std::optional<std::vector<LineAction>>>>(j, "action"));
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
        x.set_line_class(get_stack_optional<ClassEnum>(j, "class"));
        x.set_link(get_stack_optional<std::string>(j, "link"));
        x.set_redraw(get_stack_optional<Redraw>(j, "redraw"));
        x.set_showmode(get_stack_optional<int64_t>(j, "showmode"));
        x.set_type(get_untyped(j, "type"));
        x.set_replay(get_stack_optional<Replay>(j, "replay"));
        x.set_trans(get_stack_optional<Trans>(j, "trans"));
        x.set_update(get_stack_optional<Update>(j, "update"));
    }

    inline void to_json(json & j, PurpleLine const & x) {
        j = json::object();
        j["action"] = x.get_action();
        j["name"] = x.get_name();
        j["class"] = x.get_line_class();
        j["link"] = x.get_link();
        j["redraw"] = x.get_redraw();
        j["showmode"] = x.get_showmode();
        j["type"] = x.get_type();
        j["replay"] = x.get_replay();
        j["trans"] = x.get_trans();
        j["update"] = x.get_update();
    }

    inline void from_json(json const & j, DatumClass& x) {
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
        x.set_replay(get_stack_optional<Replay>(j, "replay"));
        x.set_update(get_stack_optional<Update>(j, "update"));
        x.set_action(get_stack_optional<std::vector<std::vector<DatumAction>>>(j, "action"));
        x.set_datum_class(get_stack_optional<ClassEnum>(j, "class"));
        x.set_redraw(get_stack_optional<Redraw>(j, "redraw"));
        x.set_showmode(get_stack_optional<int64_t>(j, "showmode"));
        x.set_type(get_untyped(j, "type"));
    }

    inline void to_json(json & j, DatumClass const & x) {
        j = json::object();
        j["name"] = x.get_name();
        j["replay"] = x.get_replay();
        j["update"] = x.get_update();
        j["action"] = x.get_action();
        j["class"] = x.get_datum_class();
        j["redraw"] = x.get_redraw();
        j["showmode"] = x.get_showmode();
        j["type"] = x.get_type();
    }

    inline void from_json(json const & j, LineEnv& x) {
        x.set_action(get_stack_optional<std::vector<nlohmann::json>>(j, "action"));
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
    }

    inline void to_json(json & j, LineEnv const & x) {
        j = json::object();
        j["action"] = x.get_action();
        j["name"] = x.get_name();
    }

    inline void from_json(json const & j, List& x) {
        x.set_mode(get_stack_optional<int64_t>(j, "mode"));
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
    }

    inline void to_json(json & j, List const & x) {
        j = json::object();
        j["mode"] = x.get_mode();
        j["name"] = x.get_name();
    }

    inline void from_json(json const & j, FluffyLine& x) {
        x.set_data(get_stack_optional<std::vector<std::vector<DatumElement>>>(j, "data"));
        x.set_env(get_stack_optional<LineEnv>(j, "env"));
        x.set_list(get_stack_optional<std::vector<std::optional<List>>>(j, "list"));
        x.set_wait(get_stack_optional<int64_t>(j, "wait"));
        x.set_method(get_stack_optional<Method>(j, "method"));
        x.set_time(get_stack_optional<std::variant<int64_t, std::string>>(j, "time"));
    }

    inline void to_json(json & j, FluffyLine const & x) {
        j = json::object();
        j["data"] = x.get_data();
        j["env"] = x.get_env();
        j["list"] = x.get_list();
        j["wait"] = x.get_wait();
        j["method"] = x.get_method();
        j["time"] = x.get_time();
    }

    inline void from_json(json const & j, Next& x) {
        x.set_evalstorage(get_stack_optional<Evalstorage>(j, "evalstorage"));
        x.set_type(get_stack_optional<int64_t>(j, "type"));
        x.set_storage(get_stack_optional<StorageEnum>(j, "storage"));
        x.set_target(get_stack_optional<std::string>(j, "target"));
        x.set_eval(get_stack_optional<Eval>(j, "eval"));
    }

    inline void to_json(json & j, Next const & x) {
        j = json::object();
        j["evalstorage"] = x.get_evalstorage();
        j["type"] = x.get_type();
        j["storage"] = x.get_storage();
        j["target"] = x.get_target();
        j["eval"] = x.get_eval();
    }

    inline void from_json(json const & j, PurpleText& x) {
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
        x.set_pan(get_stack_optional<int64_t>(j, "pan"));
        x.set_type(get_stack_optional<int64_t>(j, "type"));
        x.set_voice(get_stack_optional<std::string>(j, "voice"));
    }

    inline void to_json(json & j, PurpleText const & x) {
        j = json::object();
        j["name"] = x.get_name();
        j["pan"] = x.get_pan();
        j["type"] = x.get_type();
        j["voice"] = x.get_voice();
    }

    inline void from_json(json const & j, TextEnv& x) {
        x.set_name(get_stack_optional<ClassEnum>(j, "name"));
    }

    inline void to_json(json & j, TextEnv const & x) {
        j = json::object();
        j["name"] = x.get_name();
    }

    inline void from_json(json const & j, FluffyText& x) {
        x.set_data(get_stack_optional<std::vector<std::vector<DatumElement>>>(j, "data"));
        x.set_env(get_stack_optional<TextEnv>(j, "env"));
    }

    inline void to_json(json & j, FluffyText const & x) {
        j = json::object();
        j["data"] = x.get_data();
        j["env"] = x.get_env();
    }

    inline void from_json(json const & j, Scene& x) {
        x.set_first_line(get_stack_optional<int64_t>(j, "firstLine"));
        x.set_label(get_stack_optional<std::string>(j, "label"));
        x.set_lines(get_stack_optional<std::vector<SceneLine>>(j, "lines"));
        x.set_nexts(get_stack_optional<std::vector<Next>>(j, "nexts"));
        x.set_sp_count(get_stack_optional<int64_t>(j, "spCount"));
        x.set_title(get_stack_optional<std::vector<Title>>(j, "title"));
        x.set_version(get_stack_optional<int64_t>(j, "version"));
        x.set_texts(get_stack_optional<std::vector<std::vector<SceneText>>>(j, "texts"));
    }

    inline void to_json(json & j, Scene const & x) {
        j = json::object();
        j["firstLine"] = x.get_first_line();
        j["label"] = x.get_label();
        j["lines"] = x.get_lines();
        j["nexts"] = x.get_nexts();
        j["spCount"] = x.get_sp_count();
        j["title"] = x.get_title();
        j["version"] = x.get_version();
        j["texts"] = x.get_texts();
    }

    inline void from_json(json const & j, ScnJson& x) {
        x.set_hash(get_stack_optional<std::string>(j, "hash"));
        x.set_languages(get_stack_optional<std::vector<std::string>>(j, "languages"));
        x.set_llmap(get_stack_optional<std::vector<Llmap>>(j, "llmap"));
        x.set_name(get_stack_optional<std::string>(j, "name"));
        x.set_outlines(get_stack_optional<std::vector<nlohmann::json>>(j, "outlines"));
        x.set_scenes(get_stack_optional<std::vector<Scene>>(j, "scenes"));
    }

    inline void to_json(json & j, ScnJson const & x) {
        j = json::object();
        j["hash"] = x.get_hash();
        j["languages"] = x.get_languages();
        j["llmap"] = x.get_llmap();
        j["name"] = x.get_name();
        j["outlines"] = x.get_outlines();
        j["scenes"] = x.get_scenes();
    }

    inline void from_json(json const & j, StorageEnum & x) {
        if (j == "bsis_ev_001.txt") x = StorageEnum::BSIS_EV_001_TXT;
        else if (j == "gameloop.txt") x = StorageEnum::GAMELOOP_TXT;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, StorageEnum const & x) {
        switch (x) {
            case StorageEnum::BSIS_EV_001_TXT: j = "bsis_ev_001.txt"; break;
            case StorageEnum::GAMELOOP_TXT: j = "gameloop.txt"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"StorageEnum\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, ClassEnum & x) {
        if (j == "bg") x = ClassEnum::BG;
        else if (j == "bgm") x = ClassEnum::BGM;
        else if (j == "character") x = ClassEnum::CHARACTER;
        else if (j == "del") x = ClassEnum::DEL;
        else if (j == "あすみ") x = ClassEnum::EMPTY;
        else if (j == "env") x = ClassEnum::ENV;
        else if (j == "init") x = ClassEnum::INIT;
        else if (j == "loopse") x = ClassEnum::LOOPSE;
        else if (j == "lse1") x = ClassEnum::LSE1;
        else if (j == "new") x = ClassEnum::NEW;
        else if (j == "stage") x = ClassEnum::STAGE;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, ClassEnum const & x) {
        switch (x) {
            case ClassEnum::BG: j = "bg"; break;
            case ClassEnum::BGM: j = "bgm"; break;
            case ClassEnum::CHARACTER: j = "character"; break;
            case ClassEnum::DEL: j = "del"; break;
            case ClassEnum::EMPTY: j = "あすみ"; break;
            case ClassEnum::ENV: j = "env"; break;
            case ClassEnum::INIT: j = "init"; break;
            case ClassEnum::LOOPSE: j = "loopse"; break;
            case ClassEnum::LSE1: j = "lse1"; break;
            case ClassEnum::NEW: j = "new"; break;
            case ClassEnum::STAGE: j = "stage"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"ClassEnum\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Handler & x) {
        if (j == "loop") x = Handler::LOOP;
        else if (j == "MoveAction") x = Handler::MOVE_ACTION;
        else if (j == "SinAction") x = Handler::SIN_ACTION;
        else if (j == "wait") x = Handler::WAIT;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Handler const & x) {
        switch (x) {
            case Handler::LOOP: j = "loop"; break;
            case Handler::MOVE_ACTION: j = "MoveAction"; break;
            case Handler::SIN_ACTION: j = "SinAction"; break;
            case Handler::WAIT: j = "wait"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Handler\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, ActionEnum & x) {
        if (j == "") x = ActionEnum::EMPTY;
        else if (j == "fade") x = ActionEnum::FADE;
        else if (j == "leveloffset") x = ActionEnum::LEVELOFFSET;
        else if (j == "#0,-370,10,-583,20,-1040,0,0,0,0,0,0,0,0,0,0,0,0,0,0") x = ActionEnum::THE_03701058320104000000000000000;
        else if (j == "#0,-380,10,-579,20,-1032,0,0,0,0,0,0,0,0,0,0,0,0,0,0") x = ActionEnum::THE_03801057920103200000000000000;
        else if (j == "visvalue") x = ActionEnum::VISVALUE;
        else if (j == "xpos") x = ActionEnum::XPOS;
        else if (j == "ypos") x = ActionEnum::YPOS;
        else if (j == "zoomx") x = ActionEnum::ZOOMX;
        else if (j == "zoomy") x = ActionEnum::ZOOMY;
        else if (j == "zpos") x = ActionEnum::ZPOS;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, ActionEnum const & x) {
        switch (x) {
            case ActionEnum::EMPTY: j = ""; break;
            case ActionEnum::FADE: j = "fade"; break;
            case ActionEnum::LEVELOFFSET: j = "leveloffset"; break;
            case ActionEnum::THE_03701058320104000000000000000: j = "#0,-370,10,-583,20,-1040,0,0,0,0,0,0,0,0,0,0,0,0,0,0"; break;
            case ActionEnum::THE_03801057920103200000000000000: j = "#0,-380,10,-579,20,-1032,0,0,0,0,0,0,0,0,0,0,0,0,0,0"; break;
            case ActionEnum::VISVALUE: j = "visvalue"; break;
            case ActionEnum::XPOS: j = "xpos"; break;
            case ActionEnum::YPOS: j = "ypos"; break;
            case ActionEnum::ZOOMX: j = "zoomx"; break;
            case ActionEnum::ZOOMY: j = "zoomy"; break;
            case ActionEnum::ZPOS: j = "zpos"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"ActionEnum\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, File & x) {
        if (j == "bg02_昼") x = File::BG02;
        else if (j == "l_black") x = File::L_BLACK;
        else if (j == "あすみ.stand") x = File::STAND;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, File const & x) {
        switch (x) {
            case File::BG02: j = "bg02_昼"; break;
            case File::L_BLACK: j = "l_black"; break;
            case File::STAND: j = "あすみ.stand"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"File\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Dress & x) {
        if (j == "普段着") x = Dress::EMPTY;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Dress const & x) {
        switch (x) {
            case Dress::EMPTY: j = "普段着"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Dress\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Pose & x) {
        if (j == "腕開き") x = Pose::EMPTY;
        else if (j == "腕閉じ") x = Pose::POSE;
        else if (j == "腕特殊") x = Pose::PURPLE;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Pose const & x) {
        switch (x) {
            case Pose::EMPTY: j = "腕開き"; break;
            case Pose::POSE: j = "腕閉じ"; break;
            case Pose::PURPLE: j = "腕特殊"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Pose\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Filename & x) {
        if (j == "bgm10") x = Filename::BGM10;
        else if (j == "セミ") x = Filename::EMPTY;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Filename const & x) {
        switch (x) {
            case Filename::BGM10: j = "bgm10"; break;
            case Filename::EMPTY: j = "セミ"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Filename\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Method & x) {
        if (j == "crossfade") x = Method::CROSSFADE;
        else if (j == "") x = Method::EMPTY;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Method const & x) {
        switch (x) {
            case Method::CROSSFADE: j = "crossfade"; break;
            case Method::EMPTY: j = ""; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Method\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, LineEnum & x) {
        if (j == "bgm_BGM10") x = LineEnum::BGM_BGM10;
        else if (j == "cg_bg02_昼") x = LineEnum::CG_BG02;
        else if (j == "cg_あすみ.stand") x = LineEnum::CG_STAND;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, LineEnum const & x) {
        switch (x) {
            case LineEnum::BGM_BGM10: j = "bgm_BGM10"; break;
            case LineEnum::CG_BG02: j = "cg_bg02_昼"; break;
            case LineEnum::CG_STAND: j = "cg_あすみ.stand"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"LineEnum\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Eval & x) {
        if (j == "EventManager.hasSequenceScenario()") x = Eval::EVENT_MANAGER_HAS_SEQUENCE_SCENARIO;
        else if (j == "f.gameloop<0") x = Eval::F_GAMELOOP_0;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Eval const & x) {
        switch (x) {
            case Eval::EVENT_MANAGER_HAS_SEQUENCE_SCENARIO: j = "EventManager.hasSequenceScenario()"; break;
            case Eval::F_GAMELOOP_0: j = "f.gameloop<0"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Eval\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Evalstorage & x) {
        if (j == "_get_next_sequence_()") x = Evalstorage::GET_NEXT_SEQUENCE;
        else if (j == "'*□%s'.sprintf(han2zen((string)intrandom(1,10)))") x = Evalstorage::S_SPRINTF_HAN2_ZEN_STRING_INTRANDOM_110;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, Evalstorage const & x) {
        switch (x) {
            case Evalstorage::GET_NEXT_SEQUENCE: j = "_get_next_sequence_()"; break;
            case Evalstorage::S_SPRINTF_HAN2_ZEN_STRING_INTRANDOM_110: j = "'*□%s'.sprintf(han2zen((string)intrandom(1,10)))"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Evalstorage\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, TextEnum & x) {
        if (j == "お兄ちゃん") x = TextEnum::EMPTY;
        else if (j == "あすみ") x = TextEnum::TEXT;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); }
    }

    inline void to_json(json & j, TextEnum const & x) {
        switch (x) {
            case TextEnum::EMPTY: j = "お兄ちゃん"; break;
            case TextEnum::TEXT: j = "あすみ"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"TextEnum\": " + std::to_string(static_cast<int>(x)));
        }
    }

    inline void from_json(json const & j, Title & x) {
        if (j == "休息") x = Title::EMPTY;
        else if (j == "Rest") x = Title::REST;
        else if (j == "휴식") x = Title::TITLE;
        else { throw std::runtime_error("Input JSON does not conform to schema!"); } 
    }

    inline void to_json(json & j, Title const & x) {
        switch (x) {
            case Title::EMPTY: j = "休息"; break;
            case Title::REST: j = "Rest"; break;
            case Title::TITLE: j = "휴식"; break;
            default: throw std::runtime_error("Unexpected value in enumeration \"Title\": " + std::to_string(static_cast<int>(x)));
        }
    }
}
namespace nlohmann {
    inline void adl_serializer<std::variant<Scn::ClassEnum, int64_t>>::from_json(json const & j, std::variant<Scn::ClassEnum, int64_t> & x) {
        if (j.is_number_integer())
            x = j.get<int64_t>();
        else if (j.is_string())
            x = j.get<Scn::ClassEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<Scn::ClassEnum, int64_t>>::to_json(json & j, std::variant<Scn::ClassEnum, int64_t> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<Scn::ClassEnum>(x);
                break;
            case 1:
                j = std::get<int64_t>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<int64_t, std::string>>::from_json(json const & j, std::variant<int64_t, std::string> & x) {
        if (j.is_number_integer())
            x = j.get<int64_t>();
        else if (j.is_string())
            x = j.get<std::string>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<int64_t, std::string>>::to_json(json & j, std::variant<int64_t, std::string> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<int64_t>(x);
                break;
            case 1:
                j = std::get<std::string>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum>>::from_json(json const & j, std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum> & x) {
        if (j.is_number())
            x = j.get<double>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::ActionClass>>();
        else if (j.is_string())
            x = j.get<Scn::ActionEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum>>::to_json(json & j, std::variant<std::vector<Scn::ActionClass>, double, Scn::ActionEnum> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::ActionClass>>(x);
                break;
            case 1:
                j = std::get<double>(x);
                break;
            case 2:
                j = std::get<Scn::ActionEnum>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine>>::from_json(json const & j, std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine> & x) {
        if (j.is_object())
            x = j.get<Scn::PurpleLine>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::IndigoLine>>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine>>::to_json(json & j, std::variant<std::vector<Scn::IndigoLine>, Scn::PurpleLine> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::IndigoLine>>(x);
                break;
            case 1:
                j = std::get<Scn::PurpleLine>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<double, Scn::ActionEnum>>::from_json(json const & j, std::variant<double, Scn::ActionEnum> & x) {
        if (j.is_number())
            x = j.get<double>();
        else if (j.is_string())
            x = j.get<Scn::ActionEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<double, Scn::ActionEnum>>::to_json(json & j, std::variant<double, Scn::ActionEnum> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<double>(x);
                break;
            case 1:
                j = std::get<Scn::ActionEnum>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<Scn::DatumClass, Scn::ClassEnum>>::from_json(json const & j, std::variant<Scn::DatumClass, Scn::ClassEnum> & x) {
        if (j.is_object())
            x = j.get<Scn::DatumClass>();
        else if (j.is_string())
            x = j.get<Scn::ClassEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<Scn::DatumClass, Scn::ClassEnum>>::to_json(json & j, std::variant<Scn::DatumClass, Scn::ClassEnum> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<Scn::DatumClass>(x);
                break;
            case 1:
                j = std::get<Scn::ClassEnum>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string>>::from_json(json const & j, std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string> & x) {
        if (j.is_number_integer())
            x = j.get<int64_t>();
        else if (j.is_string())
            x = j.get<std::string>();
        else if (j.is_object())
            x = j.get<Scn::FluffyLine>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::StickyLine>>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string>>::to_json(json & j, std::variant<std::vector<Scn::StickyLine>, Scn::FluffyLine, int64_t, std::string> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::StickyLine>>(x);
                break;
            case 1:
                j = std::get<Scn::FluffyLine>(x);
                break;
            case 2:
                j = std::get<int64_t>(x);
                break;
            case 3:
                j = std::get<std::string>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t>>::from_json(json const & j, std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t> & x) {
        if (j.is_number_integer())
            x = j.get<int64_t>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::TentacledLine>>();
        else if (j.is_string())
            x = j.get<Scn::LineEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t>>::to_json(json & j, std::variant<std::vector<Scn::TentacledLine>, Scn::LineEnum, int64_t> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::TentacledLine>>(x);
                break;
            case 1:
                j = std::get<Scn::LineEnum>(x);
                break;
            case 2:
                j = std::get<int64_t>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::StickyText>, Scn::PurpleText>>::from_json(json const & j, std::variant<std::vector<Scn::StickyText>, Scn::PurpleText> & x) {
        if (j.is_object())
            x = j.get<Scn::PurpleText>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::StickyText>>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::StickyText>, Scn::PurpleText>>::to_json(json & j, std::variant<std::vector<Scn::StickyText>, Scn::PurpleText> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::StickyText>>(x);
                break;
            case 1:
                j = std::get<Scn::PurpleText>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText>>::from_json(json const & j, std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText> & x) {
        if (j.is_string())
            x = j.get<std::string>();
        else if (j.is_object())
            x = j.get<Scn::PurpleText>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::StickyText>>();
        else if (j.is_null())
            x = std::vector<Scn::StickyText>{};
        else throw std::runtime_error("Could not deserialise!");
    }
    
    inline void adl_serializer<std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText>>::to_json(json & j, std::variant<std::string, std::vector<Scn::StickyText>, Scn::PurpleText> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::string>(x);
            case 1:
                j = std::get<std::vector<Scn::StickyText>>(x);
                break;
            case 2:
                j = std::get<Scn::PurpleText>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }

    inline void adl_serializer<std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t>>::from_json(json const & j, std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t> & x) {
        if (j.is_number_integer())
            x = j.get<int64_t>();
        else if (j.is_object())
            x = j.get<Scn::FluffyText>();
        else if (j.is_array())
            x = j.get<std::vector<Scn::TentacledText>>();
        else if (j.is_string())
            x = j.get<Scn::TextEnum>();
        else throw std::runtime_error("Could not deserialise!");
    }

    inline void adl_serializer<std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t>>::to_json(json & j, std::variant<std::vector<Scn::TentacledText>, Scn::FluffyText, Scn::TextEnum, int64_t> const & x) {
        switch (x.index()) {
            case 0:
                j = std::get<std::vector<Scn::TentacledText>>(x);
                break;
            case 1:
                j = std::get<Scn::FluffyText>(x);
                break;
            case 2:
                j = std::get<Scn::TextEnum>(x);
                break;
            case 3:
                j = std::get<int64_t>(x);
                break;
            default: throw std::runtime_error("Input JSON does not conform to schema!");
        }
    }
}
