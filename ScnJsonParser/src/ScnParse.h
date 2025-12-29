#pragma once

#include "ScnJson.h"
using namespace Scn;

struct Dialogue {
    std::string lang;
    std::string speaker;
    std::string texts;
};

struct DialogueBlock {
    std::string Speaker;
    std::vector<Dialogue> trans;
    std::string VoiceID;
};

class Scenario {
public:
    std::string ScenarioName;
    std::vector<DialogueBlock> DialogueText;
};

class ParseJson {
public:
    bool LoadScnFile(const std::string &filepath);
    void DumpTxt();
    std::shared_ptr<Scenario> getScenario() const { return scenario; }

private:
    std::shared_ptr<Scenario> scenario = std::make_shared<Scenario>();
    std::vector<std::string> languages;


    //Deserialize all
    void CleanText(const Scene &scene);
    void ParseScn(const ScnJson &scn);

    //Deserialize texts
    // void CleanText(const json &jsonsc);
    // void ParseScn(const json &j);
};