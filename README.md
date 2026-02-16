# gitaegis

gitaegis helps developers maintain clean and secure repositories by scanning source code for potential API keys, tokens, or sensitive data before committing.
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
curl -L "https://github.com/steverahardjo/gitaegis/releases/latest/download/gitaegis-linux-amd64" -o /tmp/gitaegis && \
chmod +x /tmp/gitaegis && \
sudo mv /tmp/gitaegis /usr/local/bin/gitaegis


```

### Example Workflow

Run the main script:

```bash
#Initialize a repo
git init

#Run secret scan before committing
gitaegis scan . --logging


#Write gitignore to based on the pre vious run result being saved in the logging
gitaegis gitignore
#Ad and commit changes
git add . 
git commit
```
Script integrated with Git
```bash
#initialize a repo
git init 

#use git add to wrap gitaegis scan and a git add

#initialize gitaegis to be used with git through bash
git add . 

#do a git commit
git commit

```

Script integrated with Git Pre-Hook
```bash
#initialize a repo
git init 

#initialize gitaegis to be used with before a git commit as a pre-hook
gitaegis init --prehook

#add file into git
git add

#use a git  commit, if a gitaegis flag something its going to abort a git commit
git commit 

```
---

## Configuration Reference
The `aegis.config.toml` file defines runtime behavior for gitaegis’ frontend scanning engine.  
It controls logging, parser sources, output formats, and scanning filters.

### General Settings

| Key | Type | Description | Example |
|-----|------|--------------|----------|
| `logging` | `bool` | Enables or disables verbose console logging during scans. | `true` |
| `treesitter_source` | `string` | Path to the local or vendored Tree-Sitter grammar sources. | `"path/to/treesitter"` |
| `output_format` | `[]string` | Defines output formats for scan results. Supported: `json`, `txt`, `html`. | `["json", "txt"]` |
| `use_gitignore` | `bool` | If true, excludes files listed in `.gitignore` during scanning. | `true` |

---

### Filter Section

| Key | Type | Description | Example |
|-----|------|--------------|----------|
| `ent_limit` | `float64` | Minimum entropy threshold to flag suspicious strings. | `0.8` |
| `max_file_size` | `int` | Skip files larger than this (in bytes). Helps performance and avoids binary junk. | `1024` |
| `target_regex` | `map[string]string` | Map of file type identifiers to regex patterns for targeted scanning. | `{ "go" = ".*\\.go$", "js" = ".*\\.js$" }` |

---

### Example Config

```toml
logging = true
treesitter_source = "path/to/treesitter"
output_format = ["json", "txt"]
use_gitignore = true

[filter]
ent_limit = 4.0
max_file_size = 1024
target_regex = { "go" = ".*\\.go$", "js" = ".*\\.js$" }
```

## Future Release Plan
1. Ability to obfuscate the added file blob directly through the git add API
2. Do further otpimization through object pooling

## Contributing
Contributions and improvement ideas are welcome.
Way to do it: 
1. Fork the repo
2. Create a new branch, commit, and push on said branch

## License
This project is licensed under the [MIT license](LICENSE)

