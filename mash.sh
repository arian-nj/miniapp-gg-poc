#!/bin/bash

set -e

# --- Configurations --- 
export RELEASE_MODE=dev 
export BASE_URL=https://127.0.0.1:8443
START_LOCATION=$(pwd)
GODOT_VERSION="4.4.1"
GODOT_FILE="godot${GODOT_VERSION}"
GODOT_CMD="godotcmd"

GODOT_EXPORT_TEMPLATE="godot${GODOT_VERSION}_export_templte.tpz"
GODOT_TELEGRAM_WEB_EXPORT_PATH="$(pwd)/backend/server/static/game"
GODOT_PROJECT_LOCATION="$(pwd)/game/project.godot"

create_godot_cmd() {
	cd .cache
	if [[ -e "$GODOT_CMD" ]]; then
		echo "godot executable Exists"	
	else
		echo "No Godot executanble found"
		if ! [[ -e "${GODOT_FILE}.zip" ]]; then
			wget -O ${GODOT_FILE}.zip https://github.com/godotengine/godot-builds/releases/download/${GODOT_VERSION}-stable/Godot_v${GODOT_VERSION}-stable_linux.x86_64.zip
		else
			echo "not downloading ${GODOT_FILE}.zip already exist"
		fi

		unzip ${GODOT_FILE}.zip
		mv Godot_v${GODOT_VERSION}-stable_linux.x86_64  godotcmd
	fi
	cd ..
}

check_godot_export_template() {
	cd .cache
	if [[ -e "$GODOT_EXPORT_TEMPLATE" ]]; then
		echo "export template exists"	
	else
		echo "no export template exists"
		echo "expected file $(pwd)/${GODOT_EXPORT_TEMPLATE}"
		cd ..
		exit 1
	fi
	if ! [[ -e "$HOME/.local/share/godot/export_templates/${GODOT_VERSION}.stable/" ]] then
		unzip ${GODOT_EXPORT_TEMPLATE} -d ~/.local/share/godot/export_templates/${GODOT_VERSION}.stable
		mv ~/.local/share/godot/export_templates/${GODOT_VERSION}.stable/templates/* ~/.local/share/godot/export_templates/${GODOT_VERSION}.stable/
	else
		echo "template is unziped"
	fi
	cd ..
}

br_all() {
	cd ./backend/server/static/game/
	for f in $(find . -type f ! -name '*.br'); do
		brotli --best --keep "$f"
		echo "Compressed: $f"
	done
	cd $START_LOCATION
}
export_for_web() {
	echo "starting to export Web Game"
	mkdir -v -p $GODOT_TELEGRAM_WEB_EXPORT_PATH
	./.cache/godotcmd $GODOT_PROJECT_LOCATION --quiet --headless --export-release Web "${GODOT_TELEGRAM_WEB_EXPORT_PATH}/mini.html"
	# cd $GODOT_TELEGRAM_WEB_EXPORT_PATH
	# cd ..
	# echo "taring..."
	# tar -cf game.tar --owner=0 --group=0 --no-same-owner --no-same-permissions game/	
	# echo "brotli...(it can take a while)"
	# brotli -q 9 game.tar
	# rm -r ./game
	# rm ./game.tar
	echo "Telegram Web Game Exported"
}

build_frontend() {
	echo "Building frontend"
	cd ./frontend/
	npm run build
	cd ..
}

COMMAND=$1

if [ -z "COMMAND" ]; then
	echo "i want more commands"
fi

case $COMMAND in
	serve)
		echo "serve"
		cd ./backend/
		TELEGRAM_HTTP_PORT=3000 \
			go run ./cmd/main/.
		echo "The End"
		cd ..
		;;
	
	caddy)
		echo "Starting Caddy"
		sudo caddy run --config ./Caddyfile
		;;

	front)
		# cd ./backend/server/static
		# if [[ -e "game.tar.br" ]]; then 
		# 	mv ./game.tar.br ..
		# fi;
		# cd ../../..
		build_frontend
		
		# cd ./backend/telegram/static
		# if [[ -e "../game.tar.br" ]]; then 
		# 	mv ../game.tar.br .
		# fi;
		# cd ../../..
		#
		;;
	godot)
		mkdir -p .cache
		create_godot_cmd
		# check_godot_export_template
		build_frontend
		export_for_web
		br_all 
		;;

	*)
		echo "command '$COMMAND' is Unknown"
esac

exit 0
