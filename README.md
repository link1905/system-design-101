
# System Design 101

This project is designed to introduce and guide you through the essential concepts needed to become a system designer.

Explore the official site [at this link](https://link1905.github.io/system-design-101).

---

## Contribution Guide

We encourage contributions to help improve and expand this project!

Whether you're adding new contents, or refining documentation,
feel free to open discussions or submit pull requests.
Your input is invaluable!

### Environment Setup

This project is built using the [Hugo](https://gohugo.io/) framework.
To set up your environment, follow these steps:

#### Tools

Ensure the following tools are installed on your machine and properly configured:

- [Golang](https://go.dev/)  
- [Hugo](https://gohugo.io/)  
- [D2](https://d2lang.com/)  

Make sure to set up environment variables to allow the tools to be called globally via `go`, `hugo`, and `d2` commands.

#### Submodule

This project utilizes the [Hextra theme](themes/hextra/). Initialize or update the theme module by running the following commands:

- **For first-time setup:**  

  ```bash
  git submodule update --init --recursive
  ```

- **To update the module when needed:**  

  ```bash
  git submodule update --recursive --remote
  ```

#### Render Diagrams with D2

This project relies on [D2](https://d2lang.com/) for diagram creation. Since D2 diagrams aren’t natively supported, you’ll need to run a local renderer server in a separate process:  

```bash
go run tools/d2_renderer.go
```

#### Build and Preview the Project

You can build and serve the project locally using the Hugo server:

```bash
hugo server
```

Refer to the [Hugo documentation](https://gohugo.io/commands/hugo_server/) for available flags.

### Content Structure

All content files are located in the [content](content/) folder.

To ensure clean and readable documentation, please adhere to [common markdown standards](https://github.com/markdownlint/markdownlint)!
If you're using **VS Code**, consider installing the [Markdownlint extension](https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint) to streamline linting.

### Diagrams and Diagrams as Code

This project uses [D2](https://d2lang.com/) to create diagrams such as architecture, sequence, or class diagrams.

When creating new diagrams, utilize D2 syntax for better modifiability and maintainability. Use the following syntax to define diagrams as code: ` ```d2 ``` `.

Predefined architectural components can be found in [vars](data/vars.yaml). To add a new component:

1. Place the image in the [d2-icons](static/d2-icons/) folder.  
2. Define a corresponding variable in [vars.yaml](data/vars.yaml).  

### Embedding Images

For faster loading and to prevent issues, download and use images locally (ensure that they are copyright-free):

- **For images used infrequently**:  
  Create a folder specific to the topic and place the image directly within it.
  
- **For shared/reusable images**:  
  Place the images in the [images](static/images/) folder for universal access.

### Multi-Language Support

Want to contribute content in another language? Follow the comprehensive [Hextra theme instructions for multi-language support](https://imfing.github.io/hextra/docs/advanced/multi-language/) to set up translations effectively.
