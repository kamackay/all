git tag -a $1 -m "Version $1" && git remote | xargs -L1 git push --all
