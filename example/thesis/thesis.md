---
mdoc: true
theme: thesis
title: "Bachelor-/Masterarbeit"
author: "Max Mustermann"
tags: [thesis, vorlage, koma]
page:
  size: A4
  margin: 28mm 25mm 22mm 28mm
numbering:
  enabled: true
data:
  logo: "assets/logo.svg"
  university: "Otto-von-Muster-Universität Musterstadt"
  faculty: "Fakultät für Elektrotechnik und Informationstechnik"
  subtitle: "Richtlinien zur Ausarbeitung von Bachelor- und Masterarbeiten"
  version: "Version 1.0, Juni 2026"
  submitted: "4. Juni 2026"
  birth: "geboren am 01. Januar 1970"
  birthplace: "in Musterstadt"
references:
  - key: wikipedia-de
    text: "Wikipedia, the free encyclopedia: <em>Abstract.</em> http://de.wikipedia.org/wiki/Abstract. Version: Januar 2012"
  - key: wikipedia-en
    text: "Wikipedia, the free encyclopedia: <em>Abstract (summary).</em> http://en.wikipedia.org/wiki/Abstract. Version: Januar 2012"
  - key: lanze1982
    author: "Lanze, Werner"
    title: "Das technische Manuskript – Ein Handbuch mit ausführlichen Anleitungen für Autoren und Bearbeiter"
    edition: "3"
    publisher: "Essen: Vulkan-Verlag"
    year: "1982"
    isbn: "978-3802781223"
  - key: hering2009
    author: "Hering, Lutz; Hering, Heike; Heyne, Klaus-Geert"
    title: "Technische Berichte"
    edition: "6"
    publisher: "Wiesbaden: Vieweg+Teubner"
    year: "2009"
    isbn: "978-3834805713"
---

<!-- The title page is rendered by the theme from the frontmatter `data`
     fields above (logo, university, faculty, subtitle, version, submitted,
     birth, birthplace) plus `title`/`author`. Nothing to write here. -->

<div class="matter-roman">

<section>

# Kurzreferat {.unnumbered .notoc}

Ein *Kurzreferat* ist eine prägnante Inhaltsangabe, ein Abriss ohne
Interpretation und Wertung einer wissenschaftlichen Arbeit. Kurzreferate können
in vieler Hinsicht verwendet werden, z. B. zur Feststellung von Relevanz: Es
soll schnell und exakt zu erkennen sein, ob das Dokument für die Fragestellung
relevant ist und der Leser das Originaldokument noch lesen muss [@wikipedia-de].

# Abstract {.unnumbered .notoc}

An *abstract* is a brief summary of a research article, thesis, review,
conference proceeding or any in-depth analysis of a particular subject, and is
often used to help the reader quickly ascertain the paper's purpose. When used,
an abstract always appears at the beginning of a manuscript [@wikipedia-en].

</section>

<section>

# Text der Aufgabenstellung im Original {.unnumbered .notoc}

*(Hier wird die im Original gestellte Aufgabenstellung eingefügt.)*

</section>

<section>

# Eidesstattliche Erklärung {.unnumbered .notoc}

Hiermit versichere ich, die vorliegende Arbeit selbstständig und unter
ausschließlicher Verwendung der angegebenen Literatur und Hilfsmittel erstellt
zu haben.

Die Arbeit wurde bisher in gleicher oder ähnlicher Form keiner anderen
Prüfungsbehörde vorgelegt und auch nicht veröffentlicht.

Musterstadt, den {{.Data.submitted}}

</section>

</div>

<div class="matter-main">

<div class="frontlists">

<section>

# Inhaltsverzeichnis {.unnumbered .notoc}

:::toc
depth: 3
:::

</section>

<section>

# Symbolverzeichnis {.unnumbered .notoc}

<dl class="entries"><dt>$a,\ A$</dt><dd>Skalar, auch komplexwertig</dd><dt>$\vec{a},\ \vec{A}$</dt><dd>Vektor, auch komplexwertig</dd><dt>$\eta$</dt><dd>Wirkungsgrad</dd><dt>$\kappa$</dt><dd>Leitfähigkeit</dd></dl>

</section>

<section>

# Abkürzungsverzeichnis {.unnumbered .notoc}

<dl class="entries abbr"><dt>DIN</dt><dd>Deutsches Institut für Normung</dd><dt>ISO</dt><dd>Internationale Organisation für Normung, engl. International Organization for Standardization</dd><dt>PDF</dt><dd>(trans)portables Dokumentenformat, engl. Portable Document Format</dd></dl>

</section>

<section>

# Abbildungsverzeichnis {.unnumbered .notoc}

<nav class="lof">
<a href="#fig-spannung"><span class="t">2.1&emsp;Harmonischer Zeitverlauf einer Spannung mit einer Frequenz von 50&nbsp;Hz</span></a>
<a href="#fig-neben"><span class="t">2.2&emsp;Zwei Unterabbildungen nebeneinander</span></a>
</nav>

</section>

<section>

# Tabellenverzeichnis {.unnumbered .notoc}

<nav class="lot">
<a href="#tab-raender"><span class="t">2.1&emsp;Einzuhaltende Seitenränder bei der Erstellung von Bachelor- und Masterarbeiten</span></a>
</nav>

</section>

</div>

<div class="mainmatter">

# Einleitung

Die Einleitung soll den Leser mit dem behandelten Problem bekanntmachen und das
Ziel und die Bedeutung der Arbeit aufzeigen. Weitere mögliche Inhalte einer
Einleitung sind in Abschnitt&nbsp;2.2.1 auf Seite&nbsp;<a class="pageref" href="#einleitung-1"></a> beschrieben.

# Richtlinien zur Ausarbeitung

Schriftliche Ausarbeitungen zu Bachelor-, Master- und anderen wissenschaftlichen
Arbeiten sollen nach gewissen für technisch-wissenschaftliche Berichte bewährten
Regeln aufgebaut und gestaltet sein, um den Leser in klarer Form über das
behandelte Thema zu informieren [@lanze1982][@hering2009].

## Aufbau des Berichtes

Berichte zu Bachelor- und Masterarbeiten sollen die nachstehenden Teile in der
angegebenen Reihenfolge enthalten:

- Titelblatt,
- Kurzreferat (deutsch und englisch),
- Erklärung über die Selbständigkeit der Arbeit,
- Inhaltsverzeichnis,
- Formelzeichenliste, Abkürzungs-, Abbildungs- und Tabellenverzeichnis,
- Text des Berichtes mit Einleitung, Hauptteil und Zusammenfassung,
- Literaturverzeichnis sowie Anhänge.

## Inhaltliche Gestaltung des Textes

### Einleitung

Die Einleitung soll den Leser mit dem behandelten Problem bekanntmachen und
beschreibt die Art der gestellten Aufgabe, den Erkenntnisstand, von dem die
Arbeit ausgeht, sowie das Ziel und die Bedeutung der Arbeit.

### Hauptteil

Der Hauptteil enthält eine vollständige Beschreibung der Problemlösung im
Detail. Wesentlicher Gesichtspunkt für die Gestaltung ist die wirksame
Weitergabe der geschriebenen Information an einen fachlich gebildeten Leser, der
nicht unmittelbar mit dem behandelten Problem vertraut ist.

### Zusammenfassung

Die Zusammenfassung als letzter Abschnitt des Textes enthält klare und kritische
Aussagen über die Ergebnisse der Arbeit und ihre Bedeutung sowie die Grenzen der
Gültigkeit.

## Formale Gestaltung des Berichtes

### Äußere Form

Berichte zu Bachelor- und Masterarbeiten sind im Format DIN&nbsp;A4 anzufertigen.
Der Text ist einseitig bei 1½-fachem Zeilenabstand zu schreiben. Bei Textseiten
sind folgende Randbreiten einzuhalten: unten 2,0&nbsp;cm und sonst 2,5&nbsp;cm.

### Formale Textgestaltung

Die Gliederung des Textes ist durch Nummerierung der Abschnitte nach dem
Dezimalnummernsystem zu kennzeichnen. Gleichungen und Formeln sind möglichst auf
eigene Zeilen zu schreiben und am rechten Rand fortlaufend zu nummerieren. Als
Beispiel dient der Satz des Pythagoras

$$
a^2 + b^2 = c^2 \tag{2.1}
$$

den man in Richtung von $a$ oder $b$ umstellen kann.

### Tabellen

Zahlentafeln oder Zusammenstellungen von Daten in Tabellenform sind fortlaufend
zu nummerieren und zu bezeichnen (siehe Tabelle&nbsp;2.1). Die Bezeichnung einer
Tabelle steht oberhalb der Tabelle.

<div class="tablefig" id="tab-raender">
<div class="tabcaption">Einzuhaltende Seitenränder bei der Erstellung von Bachelor- und Masterarbeiten</div>

| Position | Seitenrand (in cm) |
| :------- | :----------------: |
| links    | 2,5                |
| rechts   | 2,5                |
| oben     | 2,5                |
| unten    | 2,0                |

</div>

### Bilder

Bilder und grafische Darstellungen aller Art sind fortlaufend zu nummerieren und
zu bezeichnen (siehe Abbildung&nbsp;2.1). Bildunterschriften sollen
selbsterklärend sein.

<figure id="fig-spannung">
<img src="assets/plot.svg" alt="Spannungsverlauf">
<figcaption>Harmonischer Zeitverlauf einer Spannung mit einer Frequenz von 50&nbsp;Hz und einem Effektivwert von 230&nbsp;V</figcaption>
</figure>

Mehrere zusammengehörige Abbildungen sollten in Unterabbildungen nebeneinander
oder untereinander gesetzt werden (siehe Abbildung&nbsp;2.2).

<figure id="fig-neben">
<div class="subfigures">
<div class="subfig"><img src="assets/plot.svg" alt="links"><div class="subcap">(a) linke Unterabbildung</div></div>
<div class="subfig"><img src="assets/plot.svg" alt="rechts"><div class="subcap">(b) rechte Unterabbildung</div></div>
</div>
<figcaption>Zwei Unterabbildungen nebeneinander</figcaption>
</figure>

### Listen

Listen und Einzelnachweise können unnummeriert sein:

- foo
- bar
- foobar

Listen können auch als Aufzählungen nummeriert werden:

1. foo
2. bar
3. foobar

## Weiterführende Literatur

Die wichtigsten DIN-Normen sind in einschlägigen Taschenbüchern zusammengefasst.
Maßgebend ist jeweils die neueste gültige Ausgabe eines Normblattes.

# Zusammenfassung

Die Zusammenfassung ist der letzte Abschnitt des Textes und fasst die Ergebnisse
der Arbeit zusammen (siehe auch Abschnitt&nbsp;2.2.3 auf
Seite&nbsp;<a class="pageref" href="#zusammenfassung"></a>).

</div>

<div class="backmatter">

# Literaturverzeichnis {.unnumbered}

:::bibliography
:::

</div>

<div class="appendix">

# Diagramme {.appendix}

Mögliche Inhalte eines Anhangs sowie dessen formale Gestaltung sind in
Abschnitt&nbsp;2.3 auf Seite&nbsp;<a class="pageref" href="#formale-gestaltung-des-berichtes"></a> näher beschrieben.

# Software für die Benutzung des LaTeX-Textsatzsystems

## Microsoft Windows

Für Microsoft Windows wird die Installation einer LaTeX-Distribution wie MiKTeX
sowie eines Editors empfohlen.

## Linux

Alle nötigen Programme sind typischerweise schon vorinstalliert bzw. lassen sich
leicht über den Paketmanager nachinstallieren.

# Checklisten

## Sprache

- [ ] eine automatische Rechtschreibprüfung wurde benutzt
- [ ] ein Kommilitone hat den Text korrekturgelesen
- [ ] der Text ist im Passiv geschrieben

## Formelsatz

- [ ] jede abgesetzte Gleichung ist horizontal zentriert und fortlaufend nummeriert
- [ ] alle Einheiten sind mit aufrechten Buchstaben geschrieben

</div>

</div>
