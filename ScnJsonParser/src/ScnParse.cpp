#include "ScnParse.h"
#include <iostream>
#include <fstream>

using json = nlohmann::json;


void ParseJson::CleanText(const Scene& scene)
{
    if(!scene.get_texts().has_value()) 
    {
        std::cout << "Read texts Error" << std::endl;
        return;
    }

    for(const auto& textArray : *scene.get_texts())   //std::vector<SceneText>
    {
        for(const auto& row : textArray)   //SceneText
        {
            if(!row.has_value()) continue;

            if(auto lineptr = std::get_if<std::vector<TentacledText>>(&(*row)))
            {
                const auto& line = *lineptr;
                DialogueBlock block;

                //std::string , std::vector<StickyText>                                                                           , PurpleText
                //["お兄ちゃん", [[null,"「あすみ、大切な話がある」",13],["Brother","\"Asumi, I have something important to say.\"",43]], {name, pan, type, voice}
                if(auto nameptr = std::get_if<std::string>(&line[0]))
                {
                    block.Speaker = *nameptr;
                }

                if(auto diaptr = std::get_if<std::vector<StickyText>>(&line[1]))
                {
                    for(size_t i = 0; i < diaptr -> size(); i++)
                    {
                        const auto& transline = (*diaptr)[i];
                        Dialogue dialogue;

                        dialogue.lang = languages[i];

                        if(transline[0].has_value())
                        {
                            if(auto spk = std::get_if<std::string>(&(*transline[0])))
                                dialogue.speaker = *spk;
                        }
                        if(transline[1].has_value())
                        {
                            if(auto txt = std::get_if<std::string>(&(*transline[1])))
                                dialogue.texts = *txt;
                        }
                        block.trans.push_back(dialogue);
                    }
                }

                if(auto purpptr = std::get_if<PurpleText>(&line[2]))
                {
                    if(purpptr -> get_voice().has_value())
                        block.VoiceID = *purpptr -> get_voice();
                }

                scenario->DialogueText.push_back(block);
            }
        }

    }

}

void ParseJson::ParseScn(const ScnJson &scn)
{
    if(!scn.get_name().has_value() && !scn.get_scenes().has_value() && !scn.get_languages().has_value())
    {
        std::cout << "Read scn Error" << std::endl;
        return;
    }

    languages = *scn.get_languages();
    scenario -> ScenarioName = *scn.get_name();

    for(const auto& scene : *scn.get_scenes())
    {
        CleanText(scene);
    }
}


void ParseJson::DumpTxt()
{
    std::string filename = scenario -> ScenarioName + ".txt";
    std::ofstream out(filename);

    for(const auto& block : scenario -> DialogueText)
    {
        out << "Speaker: " << block.Speaker << '\n';
        for(const auto& dlg : block.trans)
        {
            out << "[" + dlg.lang + "]" << dlg.speaker << ": " << dlg.texts << '\n';
        }
        if(!block.VoiceID.empty())
            out << "Voice: " << block.VoiceID << '\n';
        out << '\n';
    }

    out.close();
    std::cout << "successfully dumped " << filename << std::endl;
}



bool ParseJson::LoadScnFile(const std::string& filepath)
{
    std::ifstream file(filepath);
    if(!file.is_open()) return false;

    try
    {
        json j;
        file >> j;
        ScnJson scn = j.get<ScnJson>();
        ParseScn(scn);
        return true;
    }
    catch(std::exception& e)
    {
        std::cout << "Parse Error: " << e.what() << std::endl;
        return false;
    }
}


int main(int argc, char* argv[]) 
{
    ParseJson parse;
    if(parse.LoadScnFile(argv[1]))
        parse.DumpTxt();
    return 0;
}