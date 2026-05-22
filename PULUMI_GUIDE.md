# Pulumi Developer Guide

## Spis treści
- [Pulumi Developer Guide](#pulumi-developer-guide)
  - [Spis treści](#spis-treści)
  - [1. Podstawowe komendy](#1-podstawowe-komendy)
  - [2. Preview - sprawdź co się zmieni (BEZPIECZNE)](#2-preview---sprawdź-co-się-zmieni-bezpieczne)
  - [3. Deploy - zastosuj zmiany](#3-deploy---zastosuj-zmiany)
  - [4. Co się stanie gdy wywołasz `pulumi up`?](#4-co-się-stanie-gdy-wywołasz-pulumi-up)
    - [Obecnie (wszystkie zasoby zaimportowane i zsynchronizowane):](#obecnie-wszystkie-zasoby-zaimportowane-i-zsynchronizowane)
    - [Przykład 1: Dodajesz nową Lambda do `main.go`](#przykład-1-dodajesz-nową-lambda-do-maingo)
    - [Przykład 2: Zmieniasz `Timeout` w istniejącej Lambdzie](#przykład-2-zmieniasz-timeout-w-istniejącej-lambdzie)
    - [Przykład 3: Usuwasz zasób z kodu](#przykład-3-usuwasz-zasób-z-kodu)
  - [5. Importowanie nowych zasobów](#5-importowanie-nowych-zasobów)
  - [6. Zarządzanie secretami](#6-zarządzanie-secretami)
  - [7. Debugowanie](#7-debugowanie)
  - [8. Typowy workflow developera](#8-typowy-workflow-developera)
  - [9. Ważne zasady](#9-ważne-zasady)

---

## 1. Podstawowe komendy

```bash
# Zaloguj się do Pulumi Cloud (jeśli nie jesteś zalogowany)
pulumi login

# Wybierz stack dev (jeśli masz wiele stacków)
pulumi stack select dev

# Zobacz aktualny stan stacka
pulumi stack

# Zobacz wszystkie zasoby w stacku
pulumi stack --show-ids
```

## 2. Preview - sprawdź co się zmieni (BEZPIECZNE)

```bash
# Zobacz co Pulumi planuje zrobić (nie modyfikuje niczego)
pulumi preview

# To pokaże Ci diff między kodem a rzeczywistością:
#   ~ update   - zasób zostanie zaktualizowany (in-place)
#   + create   - nowy zasób zostanie utworzony
#   - delete   - zasób zostanie usunięty
#   +- replace - zasób zostanie zastąpiony (usunięty i stworzony od nowa)
```

## 3. Deploy - zastosuj zmiany

```bash
# Zastosuj zmiany (z potwierdzeniem)
pulumi up

# Zastosuj zmiany (bez potwierdzenia - dla CI/CD)
pulumi up --yes
```

## 4. Co się stanie gdy wywołasz `pulumi up`?

### Obecnie (wszystkie zasoby zaimportowane i zsynchronizowane):
```
Resources:
    11 unchanged
```
**Nic się nie zmieni** - Pulumi porównuje kod Go ze stanem w Pulumi Cloud i z rzeczywistością na AWS.

### Przykład 1: Dodajesz nową Lambda do `main.go`
```bash
pulumi preview
# Output: + aws:lambda:Function myNewFunction create
pulumi up
# ✅ Nowa Lambda zostanie utworzona na AWS
```

### Przykład 2: Zmieniasz `Timeout` w istniejącej Lambdzie
```bash
pulumi preview
# Output: ~ aws:lambda:Function telegramVoiceDownloader update [diff: ~timeout]
pulumi up
# ✅ Timeout zostanie zaktualizowany (in-place, bez przerwy)
```

### Przykład 3: Usuwasz zasób z kodu
```bash
pulumi preview
# Output: - aws:sns:Topic stravaNotifications delete
pulumi up
# ❌⚠️ SNS topic zostanie USUNIĘTY z AWS!
# (chyba że ma protect=true - wtedy Pulumi zablokuje)
```

## 5. Importowanie nowych zasobów

Jeśli ktoś ręcznie stworzył coś na AWS i chcesz to dodać do Pulumi:

```bash
# Import istniejącego zasobu
pulumi import aws:lambda/function:Function nazwaLogiczna ARNfunkcji

# Po imporcie musisz dodać kod do main.go
# Uruchom preview żeby sprawdzić czy kod jest zgodny
pulumi preview
```

## 6. Zarządzanie secretami

```bash
# Ustaw secret w Pulumi (szyfrowany)
pulumi config set --secret telegramToken "8629586691:..."

# Odczytaj w kodzie Go:
# config.GetSecret("telegramToken")

# Zobacz jakie są configi
pulumi config
```

## 7. Debugowanie

```bash
# Zobacz szczegóły konkretnego zasobu
pulumi stack export | jq '.deployment.resources[] | select(.type == "aws:lambda/function:Function")'

# Zobacz logi z ostatniego deployu
pulumi stack --show-urns

# Refresh stanu (jeśli ktoś ręcznie zmienił coś na AWS)
pulumi refresh
```

## 8. Typowy workflow developera

```bash
# 1. Pull najnowszy kod
git pull

# 2. Sprawdź czy twój stan jest aktualny
pulumi stack

# 3. Wprowadź zmiany w main.go

# 4. Sprawdź co się zmieni
pulumi preview

# 5. Jeśli wszystko OK - deploy
pulumi up

# 6. Commit
git add infrastructure/
git commit -m "feat: add new Lambda for ..."
```

## 9. Ważne zasady

| Zasada | Opis |
|--------|------|
| **Zawsze `preview` przed `up`** | Nigdy nie deployuj na ślepo |
| **`protect=true`** | Zasoby z `protect=true` nie mogą być przypadkiem usunięte |
| **Nie edytuj ręcznie AWS** | Jeśli zmienisz coś ręcznie, Pulumi nadpisze przy następnym `up` |
| **`pulumi refresh`** | Użyj jeśli ktoś ręcznie zmienił coś na AWS (ale lepiej tego nie robić) |
