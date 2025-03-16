# DevX CLI
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FZengineChris%2Fdevx.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FZengineChris%2Fdevx?ref=badge_shield)

DevX is a command-line tool designed to simplify the development workflow for containerized applications. It helps developers manage projects, build Docker images, and update deployments in Kubernetes environments.

## Prerequisites

- [Colima](https://github.com/abiosoft/colima) for local Kubernetes development
- Colima with Docker runtime for building and managing containers

## Installation

### Homebrew
```bash
brew install zenginechris/tap/devx
```

### Nix Flake
```bash
{
  inputs = {
    devx = {
      url = "github:zenginechris/devx";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    devx
  } @ inputs: let
    configuration = {pkgs, ...}: {
      nixpkgs.overlays = [
        (final: prev: {
          devx-cli = devx.packages.${prev.system}.default;
        })
      ];
      environment.systemPackages = [
        pkgs.devx-cli
      ];
  };
}
```

### Nix NUR
```bast
coming soon
```

## Setting Up the Environment

DevX requires a running Kubernetes cluster managed by Colima. Use the following command to start Colima with the appropriate configuration:

```bash
colima start --cpu 8 --memory 26 --disk 50 --vm-type=vz --vz-rosetta --runtime docker --kubernetes
```

This command:
- Allocates 8 CPUs and 26GB of memory
- Provides 50GB of disk space
- Uses the vz virtualization type with Rosetta translation
- Sets Docker as the runtime
- Enables Kubernetes support

## Commands

### Project Management

#### List Projects
```bash
devx list
# or
devx ls
```
Lists all configured projects showing their names and build contexts.

#### Create a New Project
```bash
devx new <project-name>
# or
devx n <project-name>
```
Creates a new project with the specified name. This command:
- Adds the project to your configuration
- Creates a default Dockerfile in the project directory

#### Set Project Context
```bash
devx context <project-name> [path]
# or
devx c <project-name> [path]
```
Sets the build context for a project. If no path is specified, it uses the current directory.

#### Edit Project Configuration
```bash
devx edit <project-name> docker
# or
devx e <project-name> docker
```
Opens the project's Dockerfile in your default editor for modification.

##### For more configuration edit the project in the devx config file
```bash
cd ~.config/devx
```
Here you can find a devx.toml file that contains all of your configured projects.
The file looks a like that: 
```yaml
[[projects]]
name = 'ui'
context = '/Users/chris/github.com/project/ui' # this is the current building context that can be set by the cli
config_path = '/Users/cbartelt/.config/devx/projects/ui' # project specific configuration like the Dockerfile
contexts = [] # additional files and folders that are copied to the build context
deployment_name = 'ui' # the kubernetes deployment name. The building command will update this with the built image tag
namespace = 'default' # the namespace the deployment is in

[[projects]]
name = 'api'
context = ''
config_path = '/Users/cbartelt/.config/devx/projects/api'
contexts = []
deployment_name = 'api'
namespace = 'default'
```


#### Build and Deploy Project
```bash
devx build <project-name>
# or
devx b <project-name>
```
Builds a Docker image from the project's context and updates the deployment in your Kubernetes cluster. This command:
1. Creates a temporary build directory
2. Copies the project context and Dockerfile
3. Builds a Docker image tagged with the project name and current timestamp
4. Updates the deployment in the current Kubernetes context

## Features

- **Multi-architecture support**: Builds images for both AMD64 and ARM64 architectures when using Docker BuildX
- **Build caching**: Implements local caching to speed up subsequent builds
- **Automatic deployment**: Updates Kubernetes deployments after successful builds
- **Flexible configuration**: Supports multiple build contexts and custom Dockerfile configurations

## Configuration

DevX stores its configuration in a central location. You can manage projects and their settings through the CLI commands.

## Examples

### Create and build a new project
```bash
# Create a new project
devx new my-go-app

# Set the build context to the current directory
devx context my-go-app

# Edit the Dockerfile if needed
devx edit my-go-app docker

# Build and deploy the project
devx build my-go-app
```

### List all projects
```bash
devx list
```

## Environment Variables

- `EDITOR` or `VISUAL`: Specifies the editor to use for the `edit` command. If not set, DevX will try to use `nano`, `vim`, or `vi` in that order.

## Notes

- The CLI automatically creates a default Go application Dockerfile when creating a new project
- Project names are converted to lowercase and spaces are replaced with hyphens for use in image names and directories
- DevX requires a running Kubernetes cluster managed by Colima for deployment functionality

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FZengineChris%2Fdevx.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FZengineChris%2Fdevx?ref=badge_large)
