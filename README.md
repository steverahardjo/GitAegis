# GitAegis

GitAegis helps developers maintain clean and secure repositories by scanning source code for potential API keys, tokens, or sensitive data before committing.
It’s designed for simplicity, speed, and easy integration into existing Git workflows.
## Features
- Secret Detection: Identify API keys, tokens, and sensitive patterns using **Shannon Entrophy** and **Regular expression**.
- Utilizing Go's  goroutine concurrency for high performance.
- Lightweight integration through a **CLI interface**.
- A logging written into a JSON file


### Prerequisites

- [Git](https://git-scm.com/)
- [Bash](https://en.wikipedia.org/wiki/Bash_(Unix_shell))

### Installation

Download

```bash
```

### Example Workflow

Run the main script:

```bash
#Initialize a repo
git init

#Run secret scan before committing
gitaegis scan . --logging

#Ad and commit changes
git add . 
git commit
```
Script integrated with Git
```bash
#initialize a repo
git init

#use git add to wrap gitaegis scan and a git add
gitaegis add . 

#do a git commit
git commit
```

Or follow your project's specific instructions.

## Contributing
Contributions and improvement ideas are welcome.
Way to do it: 
1. FOrk the repo
2. Create a new branch, commit, and push on said branch
3. We will review said changes

## License
This project is licensed under the [MIT license](LICENSE)