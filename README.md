###GoBrowse - A Git Addon

Add code .gif

DISCLAIMER: So far, this code has been only lightly tested in a relatively narrow minded manner and only on a Mac. Have a look 
at the source to make sure it will work for you before running it in you project. It is intended as a tool to ease browsing 
the history of a project NOT to replace or teach ANY Git functionality. 

####What is it? 
GoBrowse is a Git extension to easily browse project history from within a sandbox in the project directory.
The goal is to ease the pain of rolling through commits, be it for learning a new project or just to grab a lost 
line of code. I
 
####How to Install?

#####For Mac
Get the source with ```go get github.com/cbergoon/git-gobrowse```

Build the source with ```go install github.com/cbergoon/git-gobrowse```

Move to a directory on your path with ```mv $GOPATH/bin/git-gobrowse /usr/local/bin/git-gobrowse```

####Usage


####Todo
1) Implement 'help' command
4) Implement 'clean' command
2) Kinda messy. Could use a lot of refactoring...
3) Better testing?
