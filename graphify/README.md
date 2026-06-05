### Setup
* `brew install python@3.12 uv`
* `uv tool install graphifyy`
* `graphify install`
* `graphify install --platform opencode`
* `graphify opencode install` - for opencode always use the knoledge graph

### Commands
[List](https://github.com/safishamsi/graphify#full-command-reference)
* `/graphify ./raw --mode deep`        # more aggressive relationship extraction
* ```
  graphify hook install              # post-commit + post-checkout hooks
  graphify hook uninstall
  graphify hook status
  ```
  

### Codegraph
* `brew install node`
* `npm i -g @colbymchenry/codegraph`
* `codegraph install`

do not commit to git, rebuild locally