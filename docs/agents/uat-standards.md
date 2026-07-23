# UAT & Spezifikations-Standards

Dieses Dokument definiert die projektspezifischen Vorgaben für Akzeptanzkriterien, User Stories und automatische E2E-Tests.

## 1. Wissen (Knowledge) & Datenformate

Spezifikationen (PRDs) und Feature-Dokumente liegen unter `docs/features/` oder `.scratch/<feature-slug>/spec.md`.

### AsciiDoc / Markdown UAT Struktur
Jedes Feature muss lesbare, maschinen- und menscheninterpretierbare Akzeptanzkriterien im **Gherkin-Format** enthalten:

```gherkin
Feature: <Feature Name in Kebab-Case oder Klartext>

  @e2e
  Scenario: <Erfolgreicher Hauptpfad / UAT-Fall>
    Given <Ausgangszustand / Kontext>
    When <Aktion des Benutzers>
    Then <Erwartetes Systemverhalten>
```

### Regelsatz für Akzeptanzkriterien
1. **Vollständigkeit**: Jede User Story besitzt mindestens 1 automatisierbares Gherkin-Szenario.
2. **`@e2e`-Tagging**: Tests, die vom Playwright E2E-Runner ausgeführt werden können, erhalten zwingend das `@e2e`-Tag.
3. **Seam-Isolation**: Tests prüfen das Verhalten am höchsten verfügbaren Seam (z. B. UI-Interaktion gegen echtes/gemocktes API-Verhalten), ohne interne Implementierungsdetails zu koppeln.

## 2. Prozess-Einbindung (Skills & Agenten)

- **`/to-spec`**: Wenn `/to-spec` ausgeführt wird, übernimmt der Agent diese Gherkin-Standards für die Abschnitte `User Stories` und `Testing Decisions`.
- **`/to-tickets`**: Wenn `/to-tickets` Einzel-Tickets unter `.scratch/<feature-slug>/issues/` anlegt, werden die Akzeptanzkriterien der Tickets direkt aus den Gherkin-Szenarien abgeleitet.
- **`frontend-e2e-tester` (`/run-e2e`)**: Parst die Spezifikationsdateien nach `@e2e`-Szenarien und führt diese deterministisch mit Playwright aus.
