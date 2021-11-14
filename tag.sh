echo "package main\n\nconst VERSION = $1\n" > ./version.go && git add version.go && git commit -m "Version Increment" && git tag -a $1 -m "Version $1" && git remote | xargs -L1 -I {} git push {} $1
