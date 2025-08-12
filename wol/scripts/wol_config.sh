#!/bin/sh

source /koolshare/scripts/base.sh

BIN=/koolshare/bin/wol

fun_wol(){
    ${BIN} --if br0 "`dbus get wol_MAC`"
}

# =============================================
# for web submit
case $2 in
1)
	fun_wol
	http_response "$1"
    ;;
esac
