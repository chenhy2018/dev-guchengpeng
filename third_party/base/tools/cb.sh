#!/bin/bash

read -p "Github.com username: " uname
stty -echo
read -p "Password: " passw; echo
stty echo

repos='base io account biz ws nio io service op image fileop ffmpeg bd opencv'

for r in $repos
do
	echo -n "$r"" ..."
	l="${#r}"
	for ((i=15;$i>$l;i=$i-1)) do echo -n " ";done
	sha1_master=`curl -u "$uname"":""$passw" 'https://api.github.com/repos/qbox/'"$r"'/branches/master' 2>>/dev/null | grep 'git/trees' | sed -e 's,.*git/trees/\(.*\)\"$,\1,'`
	sha1_develop=`curl -u "$uname"":""$passw" 'https://api.github.com/repos/qbox/'"$r"'/branches/develop' 2>>/dev/null | grep 'git/trees' | sed -e 's,.*git/trees/\(.*\)\"$,\1,'`
	if [ -z "$sha1_master" ]
	then
		echo -e '\e[1;31m[error!]\e[0m'
		continue
	fi
	if [ -z "$sha1_develop" ]
	then
		echo -e '\e[1;31m[error!]\e[0m'
		continue
	fi
	if [ "$sha1_master" != "$sha1_develop" ]
	then
		echo -e '\e[1;31m[need-merge]\e[0m'
		continue
	fi
	echo '[up-to-date]'
done
