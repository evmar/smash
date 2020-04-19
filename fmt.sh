prettier=web/node_modules/.bin/prettier
exec $prettier *.md docs/*.md web/*.js web/src/**.ts web/dist/{index.html,style.css,manifest.json} "$@"
