#ifndef CPROCS_HPP_
#define CPROCS_HPP_

#include <functional>
#include <map>
#include <vector>
#include <regex>
#include <iostream>
#include <string>
#include <cstdlib>
#include "structs.hpp"
#include "json.hpp"
using json = nlohmann::json;

namespace omm {

  std::map<std::string, std::function<Returner(
    Action v,
    json cli_params,
    std::map<std::string, Variable> vars,
    std::deque<std::map<std::string, std::vector<Action>>> this_vals,
    std::string dir
  )>> cprocs = {

    //files.read
    { "files.read", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      Returner ret;
      std::vector<std::string> retNo;

      if (isDir(nDir + filename)) {

        std::vector<std::string> dirs = read_dir(nDir + filename);

        Action dir_arr = arrayVal;

        unsigned long long i = 0;
        for (std::string curD : dirs) {

          Action curDirOmm = ommtypes::to_string(curD);
          dir_arr.Hash_Values[std::to_string(i)] = { curDirOmm };

          ++i;
        }

        ret.value = retNo;
        ret.variables = vars;
        ret.exp = dir_arr;
        ret.type = "expression";

        return ret;

      } else {
        std::string content = readfile(&(nDir + filename)[0]);

        std::vector<std::string> retNo;

        Action contentJ = strPlaceholder;

        contentJ.ExpStr = {content};

        //make the hash values of the std::string
        for (unsigned long long o = 0; o < content.length(); o++) {
          Action curChar = strPlaceholder;

          curChar.ExpStr = {
            std::string(1, content[o])
          };

          contentJ.Hash_Values[std::to_string(o)] = { curChar };
        }

        Action returner = contentJ;

        if (v.SubCall.size() > 0) {

          Action callProcessParser = v;

          bool isProc = v.SubCall[0].IsProc;

          callProcessParser.Indexes = v.SubCall[0].Indexes;
          callProcessParser.Args = v.SubCall[0].Args;
          callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

          returner = processParser(returner, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;

        }

        ret.value = retNo;
        ret.variables = vars;
        ret.exp = returner;
        ret.type = "expression";

        return ret;
      }

    } },
    { "files.write", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {
      //get both arguments and parse them
      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];
      Action content = parser(v.Args[1], cli_params, vars, false, true, this_vals, dir).exp;

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      std::string contentstr = content.ExpStr[0];
      writefile(&(nDir + filename)[0], &contentstr[0]);

      Action returner = content;

      if (v.SubCall.size() > 0) {

        Action callProcessParser = v;

        bool isProc = v.SubCall[0].IsProc;

        callProcessParser.Indexes = v.SubCall[0].Indexes;
        callProcessParser.Args = v.SubCall[0].Args;
        callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

        returner = processParser(returner, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;
      }

      Returner ret;
      std::vector<std::string> retNo;

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = returner;
      ret.type = "expression";

      return ret;

    } },
    { "files.remove", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      deletefile(nDir + filename);

      std::vector<std::string> retNo;

      return Returner{ retNo, vars, falseyVal, "expression" };

    } },
    { "files.exists", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      //if it is not a directory and not a file, it does not exist
      bool exists = !(!isDir(nDir + filename) && !isFile(nDir + filename));

      Action returner = exists ? trueRet : falseRet;

      if (v.SubCall.size() > 0) {

        Action callProcessParser = v;

        bool isProc = v.SubCall[0].IsProc;

        callProcessParser.Indexes = v.SubCall[0].Indexes;
        callProcessParser.Args = v.SubCall[0].Args;
        callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

        returner = processParser(returner, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;

      }

      Returner ret;
      std::vector<std::string> retNo;

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = returner;
      ret.type = "expression";

      return ret;

    } },
    { "files.isFile", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      bool isFileVal = isFile(nDir + filename);

      Action returner = isFileVal ? trueRet : falseRet;

      if (v.SubCall.size() > 0) {

        Action callProcessParser = v;

        bool isProc = v.SubCall[0].IsProc;

        callProcessParser.Indexes = v.SubCall[0].Indexes;
        callProcessParser.Args = v.SubCall[0].Args;
        callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

        returner = processParser(returner, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;

      }

      Returner ret;
      std::vector<std::string> retNo;

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = returner;
      ret.type = "expression";

      return ret;

    } },
    { "files.isDir", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string filename = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      smatch match;

      //see if the filename is absolute
      std::regex pat("^[a-zA-Z]:");
      bool isOnDrive = std::regex_search(filename, match, pat);

      std::string nDir = isOnDrive ? "" : cli_params["Files"]["DIR"];

      bool isDirVal = isDir(nDir + filename);

      Action returner = isDirVal ? trueRet : falseRet;

      if (v.SubCall.size() > 0) {

        Action callProcessParser = v;

        bool isProc = v.SubCall[0].IsProc;

        callProcessParser.Indexes = v.SubCall[0].Indexes;
        callProcessParser.Args = v.SubCall[0].Args;
        callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

        returner = processParser(returner, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;
      }

      Returner ret;
      std::vector<std::string> retNo;

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = returner;
      ret.type = "expression";

      return ret;

    } },
    { "regex.match", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string str = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];
      std::string regstr = parser(v.Args[1], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      try {
        std::regex reg(regstr);

        smatch matcher;

        std::vector<unsigned long long> found_indexes;

        //get all matches
        for (auto it = std::sregex_iterator(str.begin(), str.end(), reg); it != std::sregex_iterator(); it++) {
          found_indexes.push_back(it->position());
        }

        Action returnerArr = arrayVal;

        char* cur = "0";

        //loop through the indexes found and store them an omm type array
        for (int i : found_indexes) {

          //store the value of the number 1
          Action indexJ = val1;

          indexJ.ExpStr[0] = to_string(i);

          returnerArr.Hash_Values[std::string(cur)] = { indexJ };
          cur = AddC(cur, "1", &cli_params.dump()[0]);
        }

        Action returnerVal = hashVal;

        if (v.SubCall.size() > 0) {

          Action callProcessParser = v;

          bool isProc = v.SubCall[0].IsProc;

          callProcessParser.Indexes = v.SubCall[0].Indexes;
          callProcessParser.Args = v.SubCall[0].Args;
          callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

          returnerVal = processParser(returnerArr, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;

        }

        Returner ret;
        std::vector<std::string> retNo;

        ret.value = retNo;
        ret.variables = vars;
        ret.exp = returnerVal;
        ret.type = "expression";

        return ret;

      } catch (std::regex_error& e) {

        //give information about the warning
        cout << "Warning during interpreting: Invalid Regular Expression: " << regstr << endl;
        cout << "Error description: " << e.what() << endl;
        cout << "Error code: " << e.code() << endl;
        cout << endl << std::string(90, '-') << "\n\n";

        Returner ret;
        std::vector<std::string> retNo;

        //if there is a regex error return undef
        ret.value = retNo;
        ret.variables = vars;
        ret.exp = falseyVal;
        ret.type = "expression";

        return ret;
      }

    } },
    { "regex.replace", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string str = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];
      std::string regstr = parser(v.Args[1], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];
      std::string replace_with = parser(v.Args[2], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      try {
        std::regex reg(regstr);

        std::string result = std::regex_replace(str, reg, replace_with);

        Action resultJ = strPlaceholder;

        resultJ.ExpStr[0] = result;

        char* cur = "0";

        for (char i : result) {

          Action indexJ = strPlaceholder;

          indexJ.ExpStr = { std::to_string(i) };

          resultJ.Hash_Values[std::string(cur)] = { indexJ };
          cur = AddC(cur, "1", &cli_params.dump()[0]);
        }

        Action retExp = resultJ;

        if (v.SubCall.size() > 0) {

          Action callProcessParser = v;

          bool isProc = v.SubCall[0].IsProc;

          callProcessParser.Indexes = v.SubCall[0].Indexes;
          callProcessParser.Args = v.SubCall[0].Args;
          callProcessParser.SubCall.erase(callProcessParser.SubCall.begin());

          retExp = processParser(resultJ, callProcessParser, cli_params, &vars, this_vals, isProc, dir).exp;

        }

        Returner ret;
        std::vector<std::string> retNo;

        ret.value = retNo;
        ret.variables = vars;
        ret.exp = retExp;
        ret.type = "expression";

        return ret;

      } catch (std::regex_error& e) {

        //give information about the warning
        cout << "Warning during interpreting: Invalid Regular Expression: " << regstr << endl;
        cout << "Error description: " << e.what() << endl;
        cout << "Error code: " << e.code() << endl;
        cout << endl << std::string(90, '-') << "\n\n";

        //if there is a regex error return undef
        Returner ret;
        std::vector<std::string> retNo;

        ret.value = retNo;
        ret.variables = vars;
        ret.exp = falseyVal;
        ret.type = "expression";

        return ret;
      }

    } },
    { "exec", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string cmd = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0]; //get the command
      std::string put_stdin = parser(v.Args[1], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0]; //get the stdin

      char* cmdC = &cmd[0];
      char* stdinC = &put_stdin[0];

      //get stdout from the cmd
      char* get_stdout = ExecCmd(cmdC, stdinC, &dir[0]);

      Returner ret;
      std::vector<std::string> retNo;

      Action stdout_str = strPlaceholder;

      stdout_str.ExpStr = { std::string(get_stdout) };

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = stdout_str;
      ret.type = "expression";

      return ret;

    } },
    { "read", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      std::string output = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp.ExpStr[0];

      cout << output;

      std::string val;
      cin >> val;

      Action omm_str = ommtypes::to_string(val);

      Returner ret;
      std::vector<std::string> retNo;

      ret.value = retNo;
      ret.variables = vars;
      ret.exp = omm_str;
      ret.type = "expression";

      return ret;

    } },
    { "typeof", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      Returner parsed = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir);

      Action exp = parsed.exp;
      Action stringval = strPlaceholder;

      stringval.ExpStr = { exp.Type };

      std::vector<std::string> noRet;

      return Returner{ noRet, vars, stringval, "expression" };

    } },
    { "ascii", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      Action parsed = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp;

      std::vector<std::string> returnNone;

      if (parsed.Type != "string") return Returner{ returnNone, vars, falseyVal, "expression" };
      else {

        std::string val = parsed.ExpStr[0];
        int first = (int) val[0];

        Action ascVal = val1;

        ascVal.ExpStr[0] = std::to_string(first);

        return Returner{returnNone, vars, ascVal, "expression"};
      }

      return Returner{returnNone, vars, falseyVal, "expression"};

    } },
    { "env", [](Action v, json cli_params, std::map<std::string, Variable> vars, std::deque<std::map<std::string, std::vector<Action>>> this_vals, std::string dir) -> Returner {

      Action parsed = parser(v.Args[0], cli_params, vars, false, true, this_vals, dir).exp;

      std::vector<std::string> returnNone;

      if (parsed.Type != "string") return Returner{ returnNone, vars, falseyVal, "expression" };
      else {

        std::string val = parsed.ExpStr[0];
        Action variable;

        const char* cvariable = std::getenv(val.c_str());

        if (cvariable != NULL) {

          variable = strPlaceholder;
          variable.ExpStr[0] = string(cvariable);

        } else variable = falseyVal;

        return Returner{returnNone, vars, variable, "expression"};
      }

      return Returner{returnNone, vars, falseyVal, "expression"};

    } }

  };

}

#endif
