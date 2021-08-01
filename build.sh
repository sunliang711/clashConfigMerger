#!/bin/bash
if [ -z "${BASH_SOURCE}" ]; then
    this=${PWD}
else
    rpath="$(readlink ${BASH_SOURCE})"
    if [ -z "$rpath" ]; then
        rpath=${BASH_SOURCE}
    elif echo "$rpath" | grep -q '^/'; then
        # absolute path
        echo
    else
        # relative path
        rpath="$(dirname ${BASH_SOURCE})/$rpath"
    fi
    this="$(cd $(dirname $rpath) && pwd)"
fi

export PATH=$PATH:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

user="${SUDO_USER:-$(whoami)}"
home="$(eval echo ~$user)"

# export TERM=xterm-256color

# Use colors, but only if connected to a terminal, and that terminal
# supports them.
if which tput >/dev/null 2>&1; then
  ncolors=$(tput colors 2>/dev/null)
fi
if [ -t 1 ] && [ -n "$ncolors" ] && [ "$ncolors" -ge 8 ]; then
    RED="$(tput setaf 1)"
    GREEN="$(tput setaf 2)"
    YELLOW="$(tput setaf 3)"
    BLUE="$(tput setaf 4)"
    CYAN="$(tput setaf 5)"
    BOLD="$(tput bold)"
    NORMAL="$(tput sgr0)"
else
    RED=""
    GREEN=""
    YELLOW=""
    CYAN=""
    BLUE=""
    BOLD=""
    NORMAL=""
fi

_err(){
    echo "$*" >&2
}

_command_exists(){
    command -v "$@" > /dev/null 2>&1
}

rootID=0

_runAsRoot(){
    local trace=0
    local subshell=0
    local nostdout=0
    local nostderr=0

    local optNum=0
    for opt in ${@};do
        case "${opt}" in
            --trace|-x)
                trace=1
                ((optNum++))
                ;;
            --subshell|-s)
                subshell=1
                ((optNum++))
                ;;
            --no-stdout)
                nostdout=1
                ((optNum++))
                ;;
            --no-stderr)
                nostderr=1
                ((optNum++))
                ;;
            *)
                break
                ;;
        esac
    done

    shift $(($optNum))
    local cmd="${*}"
    bash_c='bash -c'
    if [ "${EUID}" -ne "${rootID}" ];then
        if _command_exists sudo; then
            bash_c='sudo -E bash -c'
        elif _command_exists su; then
            bash_c='su -c'
        else
            cat >&2 <<-'EOF'
			Error: this installer needs the ability to run commands as root.
			We are unable to find either "sudo" or "su" available to make this happen.
			EOF
            exit 1
        fi
    fi

    local fullcommand="${bash_c} ${cmd}"
    if [ $nostdout -eq 1 ];then
        cmd="${cmd} >/dev/null"
    fi
    if [ $nostderr -eq 1 ];then
        cmd="${cmd} 2>/dev/null"
    fi

    if [ $subshell -eq 1 ];then
        if [ $trace -eq 1 ];then
            (set -x; ${bash_c} "${cmd}")
        else
            (${bash_c} "${cmd}")
        fi
    else
        if [ $trace -eq 1 ];then
            set -x; ${bash_c} "${cmd}";set +x;
        else
            ${bash_c} "${cmd}"
        fi
    fi
}

function _insert_path(){
    if [ -z "$1" ];then
        return
    fi
    echo -e ${PATH//:/"\n"} | grep -c "^$1$" >/dev/null 2>&1 || export PATH=$1:$PATH
}

_run(){
    local trace=0
    local subshell=0
    local nostdout=0
    local nostderr=0

    local optNum=0
    for opt in ${@};do
        case "${opt}" in
            --trace|-x)
                trace=1
                ((optNum++))
                ;;
            --subshell|-s)
                subshell=1
                ((optNum++))
                ;;
            --no-stdout)
                nostdout=1
                ((optNum++))
                ;;
            --no-stderr)
                nostderr=1
                ((optNum++))
                ;;
            *)
                break
                ;;
        esac
    done

    shift $(($optNum))
    local cmd="${*}"
    bash_c='bash -c'

    local fullcommand="${bash_c} ${cmd}"
    if [ $nostdout -eq 1 ];then
        cmd="${cmd} >/dev/null"
    fi
    if [ $nostderr -eq 1 ];then
        cmd="${cmd} 2>/dev/null"
    fi

    if [ $subshell -eq 1 ];then
        if [ $trace -eq 1 ];then
            (set -x; ${bash_c} "${cmd}")
        else
            (${bash_c} "${cmd}")
        fi
    else
        if [ $trace -eq 1 ];then
            set -x; ${bash_c} "${cmd}";set +x;
        else
            ${bash_c} "${cmd}"
        fi
    fi
}

function _root(){
    if [ ${EUID} -ne ${rootID} ];then
        echo "Requires root privilege."
        exit 1
    fi
}

ed=vi
if _command_exists vim; then
    ed=vim
fi
if _command_exists nvim; then
    ed=nvim
fi
# use ENV: editor to override
if [ -n "${editor}" ];then
    ed=${editor}
fi
###############################################################################
# write your code below (just define function[s])
# function is hidden when begin with '_'
set -e
# Correct ME
exeName="clashMerger"
# separated by space or newline,quote item if item including space
declare -a runtimeFiles=(
template.yaml
)
# FIX ME
# example: main.GitHash or packageName/path/to/hello.GitHash
gitHashPath=main.GitHash
# FIX ME
# example: main.BuildTime or packageName/path/to/hello.BuildTime
buildTimePath=main.BuildTime
# FIX ME
# example: main.BuildMachine or packageName/path/to/hello.BuildMachine
buildMachinePath=main.BuildMachine
_build(){
    local os=${1:?'missing GOOS'}
    local arch=${2:?'missing GOARCH'}
    if [ -z ${exeName} ];then
        echo "${RED}Error: exeName not set!${NORMAL}"
        exit 1
    fi
    local resultDir=binary/"${exeName}-${os}-${arch}"

    if [ ${#runtimeFiles} -eq 0 ];then
        echo "${YELLOW}Warning: runtimeFiles is empty!${NORMAL}"
    fi

    if [ ! -d ${resultDir} ];then
        mkdir -p ${resultDir}
    fi

    ldflags="-w -s"
    if [ -n "${gitHashPath}" ];then
        local gitHash="$(git rev-parse HEAD)"
        ldflags="${ldflags} -X ${gitHashPath}=${gitHash}"
    else
        echo "${YELLOW}Warning: gitHashPath is not set${NORMAL}"
    fi

    if [ -n "${buildTimePath}" ];then
        local buildTime="$(date +%FT%T)"
        ldflags="${ldflags} -X ${buildTimePath}=${buildTime}"
    else
        echo "${YELLOW}Warning: buildTimePath is not set${NORMAL}"
    fi

    if [ -n "${buildMachinePath}" ];then
        local buildMachine="$(uname -s)-$(uname -m)"
        ldflags="${ldflags} -X ${buildMachinePath}=${buildMachine}"
    else
        echo "${YELLOW}Warning: buildMachinePath is not set${NORMAL}"
    fi

    echo "${GREEN}Build ${exeName} to ${resultDir}...${NORMAL}"
    GOOS=${os} GOARCH=${arch} go build -o ${resultDir}/${exeName} -ldflags "${ldflags}" main.go && { echo "${GREEN}Build successfully.${NORMAL}"; } || { echo "${RED}Build failed${NORMAL}"; /bin/rm -rf "${resultDir}"; exit 1; }
    for f in "${runtimeFiles[@]}";do
        cp $f ${resultDir}
    done
}

build(){
    _build darwin amd64
    _build linux amd64
    _build linux arm64
}

_pack(){
    local os=${1:?'missing GOOS'}
    local arch=${2:?'missing GOARCH'}
    local resultDir="${exeName}-${os}-${arch}"

    _build $os $arch
    tar -jcvf ${resultDir}.tar.bz2 ${resultDir}
    /bin/rm -rf ${resultDir}
}

pack(){
    _pack darwin amd64
    _pack linux amd64
    _pack linux arm64
}

# write your code above
###############################################################################

em(){
    $ed $0
}

function _help(){
    cd "${this}"
    cat<<EOF2
Usage: $(basename $0) ${bold}CMD${reset}

${bold}CMD${reset}:
EOF2
    perl -lne 'print "\t$2" if /^\s*(function)?\s*(\S+)\s*\(\)\s*\{$/' $(basename ${BASH_SOURCE}) | perl -lne "print if /^\t[^_]/"
}

case "$1" in
     ""|-h|--help|help)
        _help
        ;;
    *)
        "$@"
esac
