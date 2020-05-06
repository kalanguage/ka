var readfile = require('../files/imports/read')
, rw = require('./remove_whitespace');

module.exports = (dir, lex) => {

  //loop through lex
  for (let i = 0; i < lex.length; i++)
    if (lex[i].Name == 'include') {
      var include_name = lex[i + 2].Name;

      include_name = include_name.substr(1).slice(0, -1);

      var file
      , sendDir = dir + include_name;

      //if it is a absolute directory, then do not read from the current "dir"
      if (/^[a-zA-Z]\:/.test(include_name)) {
        file = readfile(include_name);
        sendDir = include_name;
      } else file = readfile(dir + include_name);

      if (file.startsWith('Error')) {
        console.log(file);
        process.exit(1);
      }

      var lexxed = lexer(rw(JSON.parse(file)[0]), sendDir);

      let _lex = lex.slice(0, i)
      , lex_ = lex.slice(i + 3);

      lex = [..._lex, ...lexxed, ...lex_];

    }

  return lex;
};