# Exercise 03 – CI Pipeline: Matrix Builds, Linting, SonarCloud & Coverage

**Student:** Richard Auer  
**Repository:** https://github.com/uplif3/CI-CD-MCM-Auer  
**Branch:** `exercise/03-ci-pipeline`

---

## Task 1 – Matrix Builds (4 Points)

### Was wurde gemacht

Die `test`-Job-Definition in `.github/workflows/ci.yml` wurde um eine `strategy.matrix` erweitert. Statt eines einzelnen Jobs laufen nun **4 parallele Jobs** in der Matrix:

| OS | Go-Version |
|----|-----------|
| ubuntu-latest | 1.25 |
| ubuntu-latest | 1.26 |
| macos-latest | 1.25 |
| macos-latest | 1.26 |

### Relevante Änderung in `ci.yml`

```yaml
jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        go-version: ["1.25", "1.26"]
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
```

### Nachweis

GitHub Actions → Tab **Actions** → Run auswählen → unter dem `test`-Job sind 4 parallele Einträge sichtbar (ubuntu/macos × 1.25/1.26).

![Matrix Build – 4 parallele Jobs](./screenshots/01.png)

---

## Task 2 – Linting mit golangci-lint (6 Points)

### Was wurde gemacht

1. In `.golangci.yml` wurden die vier geforderten Linter aktiviert: `gofmt`, `gocyclo`, `misspell`, `gocritic`.
2. In `ci.yml` wurde ein eigenständiger `lint`-Job hinzugefügt, der `golangci/golangci-lint-action@v4` verwendet.
3. Der Lint-Job läuft mit Go 1.24, da die golangci-lint-Binary mit Go 1.24 kompiliert ist und der Einsatz von Go 1.26 zu einem Typcheck-Fehler mit `go1.26`-Build-Tag-Dateien in Dependencies führte.

### `.golangci.yml`

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - ineffassign
    - gosimple
    - gofmt
    - gocyclo
    - misspell
    - gocritic

linters-settings:
  errcheck:
    check-type-assertions: true
  gocyclo:
    min-complexity: 15
```

### Lint-Job in `ci.yml`

```yaml
lint:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: "1.24"
    - uses: golangci/golangci-lint-action@v4
      with:
        version: latest
```

### Nachweis

GitHub Actions → Run auswählen → Job `lint` ist grün und zeigt „issues found: 0".

![Lint Job – Clean Run](./screenshots/02.png)

---

## Task 3 – SonarCloud Integration (8 Points)

### Was wurde gemacht

1. `sonar-project.properties` mit Projektkey und Organisation konfiguriert.
2. In `ci.yml` wurde ein `sonarcloud`-Job ergänzt, der nach dem `test`-Job läuft, das Coverage-Artifact herunterlädt und den Scan via `SonarSource/sonarqube-scan-action@master` ausführt.
3. Das Coverage-Artifact (`coverage.out`) wird nur einmal erzeugt (ubuntu-latest, Go 1.26) und geteilt.
4. In den GitHub-Repository-Settings wurde das Secret `SONAR_TOKEN` hinterlegt.
5. Automatic Analysis auf sonarcloud.io wurde deaktiviert, da CI-Analyse und Automatic Analysis nicht gleichzeitig aktiv sein können.

### `sonar-project.properties`

```properties
sonar.projectKey=uplif3_CI-CD-MCM-Auer
sonar.organization=uplif3

sonar.sources=.
sonar.exclusions=**/*_test.go,**/vendor/**
sonar.tests=.
sonar.test.inclusions=**/*_test.go

sonar.go.coverage.reportPaths=coverage.out

sonar.sourceEncoding=UTF-8
```

### SonarCloud-Job in `ci.yml`

```yaml
sonarcloud:
  runs-on: ubuntu-latest
  needs: test
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    - name: Download coverage report
      uses: actions/download-artifact@v4
      with:
        name: coverage-report
    - name: SonarCloud Scan
      uses: SonarSource/sonarqube-scan-action@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### Nachweis

SonarCloud Dashboard: https://sonarcloud.io/project/overview?id=uplif3_CI-CD-MCM-Auer

![SonarCloud – Dashboard](./screenshots/03.png)

---

## Task 4 – Code Coverage ≥ 80% (6 Points)

### Was wurde gemacht

#### Neue Tests in `internal/store/memory_test.go`

Die bestehenden TODOs wurden implementiert und fehlende Testfälle ergänzt:

| Testfunktion | Beschreibung |
|---|---|
| `TestCreateAndGet` | Produkt anlegen, via `GetByID` zurückgeben und Name prüfen |
| `TestGetAllEmpty` | Leerer Store gibt 0 Produkte zurück |
| `TestGetByIDNotFound` | `GetByID` auf nicht existierende ID → `ErrNotFound` |
| `TestUpdate` | Produkt anlegen, aktualisieren, geänderten Namen prüfen |
| `TestUpdateNotFound` | Update auf nicht existierende ID → `ErrNotFound` |
| `TestDeleteExisting` | Produkt anlegen, löschen, `GetByID` → `ErrNotFound` |
| `TestDeleteNonExistent` | Delete auf nicht existierende ID → `ErrNotFound` |

#### Erweiterte Tests in `internal/handler/handler_test.go`

Bestehende Tests aus Exercise 02 übernommen und fehlende Edge Cases ergänzt:

| Testfunktion | Beschreibung |
|---|---|
| `TestCreateBadJSON` | POST mit ungültigem JSON → 400 |
| `TestUpdateProductNotFound` | PUT auf nicht existierende ID → 404 |
| `TestUpdateBadJSON` | PUT mit ungültigem JSON → 400 |
| `TestDeleteProductNotFound` | DELETE auf nicht existierende ID → 404 |

#### Coverage-Threshold-Check in `ci.yml`

```yaml
- name: Check coverage threshold
  if: matrix.os == 'ubuntu-latest' && matrix.go-version == '1.26'
  run: |
    grep -vE "postgres|cmd/api" coverage.out > coverage_filtered.out
    go tool cover -func=coverage_filtered.out | awk '/total:/{gsub(/%/,"",$3); if($3+0 < 80){print "Coverage "$3"% is below 80%"; exit 1}else{print "Coverage OK: "$3"%"}}'
```

> Die Postgres-Dateien (`postgres.go`, `postgres_handler.go`) sowie `cmd/api/main.go` werden aus der Coverage-Berechnung ausgeschlossen, da sie eine Datenbankverbindung voraussetzen und nicht durch Unit-Tests abdeckbar sind. Dieser Ansatz ist in Go-Projekten gängige Praxis.

### Ergebnis

```
Coverage OK: 98.8%
```

| Datei | Coverage |
|---|---|
| `internal/handler/handler.go` | 100% |
| `internal/model/product.go` | 100% |
| `internal/store/memory.go` | ~98% |
| **Gesamt (gefiltert)** | **98.8%** |

### Nachweis

GitHub Actions → Run auswählen → Job `test (ubuntu-latest, 1.26)` → Step „Check coverage threshold" → Ausgabe `Coverage OK: 98.8%`.

Das HTML-Coverage-Report (`coverage.html`) wird ebenfalls als Artifact hochgeladen und ist im Actions-Run unter **Artifacts → coverage-report** herunterladbar.

![Coverage Threshold Check – CI Ausgabe](./screenshots/04.png)
