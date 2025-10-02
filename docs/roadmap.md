# GitAegis Roadmap

## Priority 1 — Core Correctness & Detection
1. Fix bug in `Allfilters` boolean union  
   - Ensure multiple filters combine logically (AND/OR) without false positives.  
   - Foundation for reliable detection.  

2. Sample-based regex filter generator  
   - Auto-generate regex candidates from detected strings.  
   - Configurable manual regex rules via `.gitaegis.yaml`. 

3. Portable grammar loading  
   - Bundle or fetch grammars from GitHub.  
   - Allow custom grammars in config.   

---

## Priority 2 — Workflow Integration
4. Enable actual Git pre-hooks instead of bashrc  
   - `gitaegis init` should install a `.git/hooks/pre-commit` or `pre-push` hook.  
   - Prevents secrets from entering repo at commit time.  

5. Enable obfuscation directly on `git add` (no manual undo)  
   - Intercept staged files → obfuscate secrets → stage masked content.  
   - Undo step becomes optional.  

---

## Priority 3 — Config & Persistence
6. Change `blob.gitaegis` format into JSON  
   - Human-readable, portable, and easy to parse in CI/CD.  
   - Enables audit trail and integration with other tools.  

7. Review/changing mechanism for `exemptAdditor`  
   - Move exemption management into config file.  
   - Allow bulk exemptions and reversible actions.  

---

## Priority 4 — Performance & Portability
8. Parallelism (`obfuscate`, `undoObfuscate`, `calcEntropy`)  
   - Worker pool scanning for faster runs on large repos.  

---

## Priority 5 — Advanced Features
10. Hash secrets instead of masking in obfuscation  
    - Replace secrets with deterministic hashes.  
    - Allows later verification without exposing original values.  
