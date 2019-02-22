prettier=web/node_modules/.bin/prettier
exec $prettier *.md web/*.js web/src/**.ts "$@"
