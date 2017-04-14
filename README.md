### GoBrowse - A Git Addon

Add code .gif

DISCLAIMER: So far, this code has been only lightly tested with a very narrow usecase and only on macOS. Have a look 
at the source to make sure it will work for you before running it in you project. It is intended as a tool to ease browsing 
the history of a project NOT to replace or teach ANY Git functionality. 

#### What is it? 
GoBrowse is a Git add-on to easily browse project history from within a sandbox in the project directory.
The goal is to ease the pain of rolling through commits, be it for learning a new project or just to grab a lost 
line of code.
 
#### How to Install?

##### For Mac
Get the source with ```go get github.com/cbergoon/git-gobrowse```

Build the source with ```go install github.com/cbergoon/git-gobrowse```

Move to a directory on your path with ```mv $GOPATH/bin/git-gobrowse /usr/local/bin/git-gobrowse```

#### Usage
Start gobrowse by running:
```git gobrowse -i``` 
or 
```git-gobrowse -i```

This will start a REPL from which the repository can be explored with the commands in the next section. Keep in mind, 
all changes to repo state are made in the ```.sandbox``` directory. 

##### Commands
```clone```: Re-clone the current repository into the ```.sandbox``` directory.

```list```: List all commits. Notating the current commit.

```branch-list```: List all branches. 

```move \<hash>```: Move to a specific commit. 

```first```: Move to the repositories first commit. 

```last```: Move to the repositories last commit. 

```prev```: Move back one commit. 

```next```: Move forward one commit. 

```branch \<branch>```: Switch to a specified branch.

And thats it!

#### Todo
1) Implement 'help' command
4) Implement 'clean' command
2) Kinda messy. Could be a lot better with some refactoring...
3) Better testing?
4) Show +/- 10 commits on list. With ...above and below...
