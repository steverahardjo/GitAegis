# gitaegis Internals

This is a Go CLI tool developed in **Golang**.  
It is primarily used to scan folders or hook into `git add` to detect and manage potential secrets.

Methods and ideas are inspired by:
- Git pre-hooks ([Git Book](https://git-scm.com/book/ms/v2/Customizing-Git-Git-Hooks))  
- Entropy usage ([GitGuardian entropy detector](https://docs.gitguardian.com/secrets-detection/secrets-detection-engine/detectors/generics/generic_high_entropy_secret))  
- Language-agnostic parsing ([tree-sitter](https://tree-sitter.github.io/tree-sitter/))  
- Pretty print output ([GitGuardian x-api-key detector](https://docs.gitguardian.com/secrets-detection/secrets-detection-engine/detectors/generics/x_api_key))  

---

## File Structures and Functionalities

### Package: **core**
#### analyzer
Functions tying all core services together:
- `Init()`: Initialize data structures (maps, exempted files list)  
- `IterFolder()`: Iterate a folder using goroutines  
- `PrettyPrintResult()`: Pretty-print detected results from `filenameMap`  
- `IsFilenameMapEmpty()`: Check before triggering pretty print
-  `IterFolderDiff()`: Go through the folder as a git diff result instead of whole file
#### entro_parser
Per-line parser used inside analyzer and scanning actions:
- `LineFilter` struct: enables composition and inheritance for filters  
- `RegexFilter()`: basic length/regex filter  
- `calcEntropy()`: Shannon entropy calculation over runes  
- `EntropyFilter()`: wrapper returning `bool` vs entropy threshold  
- `Allfilters()`: union wrapper over multiple filters (**BUG**)  

#### file_modification
Complementary services for persistence and obfuscation:
- `SaveFileNameMap()`: persist results into `.gitaegis`  
- `LoadFileNameMap()`: load results from previous run  
- `Obfuscate()`: rewrite files with secrets hidden  
- `UndoObfuscate()`: restore files after push  
- `UpdateGitignore()`: sync detected filenames into `.gitignore`  

#### sitter
Handles **go-tree-sitter** bindings:
- `loadextMap()`: map file extensions to grammar `.so` files  
- `initGrammar()`: load grammars with C bindings  
- `createTree()`: build parse tree in memory  
- `Walkparse()`: walk syntax tree, run filters on leaf nodes  

---

### Package: **main**
#### cmd
Bridge between core services and CLI (via Cobra):
- `RootCmd()`: init data structures and root CLI (`gitaegis`)  
- `ScanCmd()`: attach `Scan()` to CLI  
- `Scan()`: pipeline (folder iteration → filters → pretty print)  
- `gitignoreCmd()`: update `.gitignore`  
- `exemptAdditor()`: add filenames to exemption list  

---

### Package: **intro**
- `tui_intro`: terminal UI intro  
- `attach_intro`: add shell configs (currently `.bashrc`)  

---

## Third-party Packages
1. [tree-sitter bindings](https://github.com/tree-sitter/go-tree-sitter)  
2. [bubbletea](https://github.com/charmbracelet/bubbletea)  
3. [cobra](https://cobra.dev/)  
4. [go-gitignore](https://github.com/sabhiram/go-gitignore)  
5. [purego](https://github.com/ebitengine/purego)  



