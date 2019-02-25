prettier=web/node_modules/.bin/prettier
exec $prettier *.md web/*.js web/src/**.ts web/dist/{index.html,manifest.json} "$@"
