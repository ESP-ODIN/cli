# ODIN CLI

Package manager d'agents IA.

---

## Installation

```sh
go build -o odin .
# ou
make build
```

Mets le binaire `odin` dans ton `$PATH`.

---

## Commandes

### `odin search`
Liste tous les agents disponibles dans le registry.

```sh
odin search
```

### `odin install <nom>`
Installe un agent depuis le registry. Clone le repo de l'agent dans `~/.odin/agents/<nom>/`, installe ses dépendances Python si besoin, et demande les clés API requises.

```sh
odin install commit-helper
```

**Option** : installer directement depuis un repo GitHub (bypass registry) :

```sh
odin install mon-agent --repo owner/repo
```

### `odin list`
Affiche les agents déjà installés sur ta machine.

```sh
odin list
```

### `odin run <nom> [args...]`
Exécute un agent installé depuis ton répertoire de travail courant. Les clés API sauvegardées sont injectées automatiquement comme variables d'environnement, ainsi que `ODIN_AGENT_NAME` et `ODIN_AGENT_DIR`.

- Tout ce qui suit le nom de l'agent (y compris les flags) est transmis tel quel à l'agent.
- stdin/stdout/stderr sont branchés directement : les agents interactifs ou longs fonctionnent normalement.
- Le code de sortie de l'agent est propagé (`odin run ... && ...` fonctionne).
- Si une clé déclarée dans `requires` n'est ni dans `~/.odin/keys.json` ni dans l'environnement, le run est refusé.

```sh
odin run commit-helper
odin run commit-helper --dry-run
```

### `odin version`
Affiche la version du CLI.

```sh
odin version
```

---

## Comment ça marche

```
odin install commit-helper
  → appel backend : GET /agents/commit-helper
  → récupère le champ "repo" (ex: ESP-ODIN/commit-helper)
  → fetch odin.toml depuis GitHub (raw)
  → clone le repo dans ~/.odin/agents/commit-helper/
  → pip install -r requirements.txt si présent
  → demande les clés API manquantes → stockées dans ~/.odin/keys.json

odin run commit-helper
  → lit ~/.odin/registry.json pour trouver l'entrypoint
  → résout le script de l'entrypoint en chemin absolu dans ~/.odin/agents/<nom>/
  → exécute depuis le cwd de l'utilisateur, en streamant stdin/stdout/stderr
  → injecte les clés de ~/.odin/keys.json en variables d'env
```

**Résolution de l'entrypoint à l'install** (`GET /api/v1/packages/{name}/{version}`) :
1. `version.manifest_toml` s'il est fourni par le backend ;
2. sinon `odin.toml` / `manifest.toml` du repo GitHub pointé par `version.repo_link` ;
3. sinon l'agent est installé sans entrypoint et `odin run` explique qu'il faut un manifest.

**Fichiers locaux :**
- `~/.odin/registry.json` — agents installés
- `~/.odin/agents/<nom>/` — fichiers de chaque agent
- `~/.odin/keys.json` — clés API (permissions 600)

---

## Format d'un agent (`odin.toml`)

Chaque agent expose un fichier `odin.toml` à la racine de son repo GitHub :

```toml
name        = "commit-helper"
version     = "1.0.0"
description = "Génère des messages de commit à partir du git diff"
entrypoint  = "python3 agent.py"
requires    = ["OPENAI_API_KEY"]
```

---

## Dev

```sh
go run .     # lancer directement
make run     # équivalent
make test
make vet
make fmt
```

**Docker (hot-reload) :**

```sh
make docker-build
make docker-dev   # monte le source, rebuild à chaque modif .go
```

---

## Récents changements

### Gestion des clés API (`internal/keys.go`)
Les clés API sont isolées dans `~/.odin/keys.json` (séparées du registry). Lors d'un `install`, l'agent déclare ses dépendances via `requires` dans `odin.toml` — le CLI demande uniquement les clés manquantes. Elles sont injectées comme variables d'environnement à chaque `run`.

### Résolution du manifest (`internal/registry.go`)
`FetchAgentByName` appelle le backend (`GET /agents/<nom>`) pour résoudre le champ `repo`, puis fetch `odin.toml` directement depuis GitHub. Le backend n'est donc qu'un **annuaire** : il retourne `{ "repo": "owner/repo", ... }`, pas le manifest complet.

### Exécution depuis le cwd (`cmd/run.go`)
`odin run` s'exécute depuis le répertoire courant de l'utilisateur (pas depuis `~/.odin/agents/`). Les chemins relatifs dans `entrypoint` sont résolus automatiquement en absolu, ce qui permet à l'agent de lire le repo git de l'utilisateur.

---

## Branchement backend — ce qui reste à faire

Le CLI est prêt. Il manque uniquement les routes backend qui servent les agents.

### Contrat API attendu

**`GET /agents`** — liste tous les agents disponibles
```json
[
  {
    "name": "commit-helper",
    "version": "1.0.0",
    "description": "Génère des messages de commit",
    "repo": "ESP-ODIN/commit-helper"
  }
]
```

**`GET /agents/:name`** — retourne un agent par son nom
```json
{
  "name": "commit-helper",
  "version": "1.0.0",
  "description": "Génère des messages de commit",
  "repo": "ESP-ODIN/commit-helper"
}
```

Le champ `repo` est **obligatoire** — c'est ce que le CLI utilise pour fetcher `odin.toml` depuis GitHub.

### Fichier à modifier côté CLI

| Fichier | Quoi changer |
|---|---|
| `internal/registry.go:16` | `RegistryURL` vaut `http://localhost:3500` par défaut — override à la compilation : `-ldflags "-X github.com/ESP-ODIN/cli/internal.RegistryURL=https://api.odin.dev"` |

Rien d'autre à toucher : `FetchAgentByName` et `FetchAgentList` sont déjà câblés sur les bonnes routes.

### Tester sans backend

```sh
odin install mon-agent --repo owner/repo
```

Bypasse le backend et fetch `odin.toml` directement depuis GitHub.
