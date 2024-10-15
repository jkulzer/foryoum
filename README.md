# Setup/Installation

## Benötigte Software

### Go Compiler
https://go.dev/doc/install

### Templ
Wenn Go installiert ist:
```
go install github.com/a-h/templ/cmd/templ@latest
```

Mehr Infos:
https://templ.guide/quick-start/installation

## Kompilierung

```
go build .
```

## Ausführung

```
./foryoum
```

## Konfiguration

Die Datenbank wird in `./sqlite.db` gespeichert. 
</br>
Der Administrator kann in `./mainPage.html` den Inhalt der Startseite beliebig gestalten.
</br>
Die Datei `./custom.html` wird auf jeder Seite ganz oben angezeigt. Dies ist vor allem für custom CSS des Administrators vorgesehen. Vorinstalliert ist ein einfaches CSS-Theme, dies kann jedoch komplett ersetzt oder angepasst werden.
</br>
Anhänge werden in `./attachments` gespeichert.

# Entwickeln

Folgende Befehle ausführen um das Programm neu zu kompilieren und auszuführen:

```
templ generate
go run .
```

## Datenbank

# Dokumentation

Gute Dokumentation für Go:
https://gobyexample.com

Dokumentation für Templ: https://templ.guide
Templ generiert das HTML

Dokumentation für HTMX: https://htmx.org/docs
HTMX ist die Frontend-Library

Dokumentation für GORM: https://gorm.io/docs
GORM kümmert sich um die Datenbankverbindung

Dokumentation für Chi: https://go-chi.io/#/pages/routing
Chi ist das Framework für den Webserver

# Projektstruktur

In `controllers` befinden sich Funktionen, welche mit den Daten interagieren, aber nicht direkt durch Useraktionen aufgerufen werden

In `db` befinden sich Funktionen zur Datenbank

In `helpers` befinden sich sonstige Funktionen, die sonst nirgendwo dazugehören

In `models` sind die Datenstrukturen, welche in der Datenbank gespeichert werden

In `routes` ist Code, welcher direkt bei Aktionen vom User aufgerufen wird

In `views` sind `.templ`-Dateien, welche das Frontend beschreiben
