# DevX
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FZengineChris%2Fdevx.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FZengineChris%2Fdevx?ref=badge_shield)


DevX is an opinionated way for local development of projects running on kubernetes. The idea is to have a tool that makes it easy to build a service and run it in a local installation of a project.

## Workflow 

First you have to creat one or more projects by runnig this command:
```shell
devx create <name>
```
You can list all of your projects with
```shell
devx list
```

Next you can set the build context of a project with: 
```shell
devx <name> context
```
This will set the current folder as the build context for the project. 
You can also set the context by providing a path to the command like so:
```shell
devx <name> context <path>
```

## Install (MacOS with brew)

### Dependencies
#### Colima 

```shell
brew intstall colima
```
#### STOW

```shell 
berw install stow
```

### Create a LimaVW with Colima 

You can customize this to match your needs. The only thing that is important is 
the docker runtime.

```shell
colima start --cpu 8 --memory 26 --disk 50 --vm-type=vz --vz-rosetta --runtime docker --kubernetes
```


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FZengineChris%2Fdevx.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2FZengineChris%2Fdevx?ref=badge_large&issueType=license)




## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FZengineChris%2Fdevx.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FZengineChris%2Fdevx?ref=badge_large)