# Markdown Image Paste (Sublime Text 4 - macOS)

Coller une image du presse-papier directement dans un fichier markdown.

Quand le presse-papier contient une image et que vous faites `Cmd+V` dans un
fichier markdown, l'image est enregistrée à côté du `.md` courant (ou dans un
sous-dossier configurable) sous le nom `<nomDuFichier>_<N>.<ext>`, et la
référence `![](chemin/relatif)` est insérée à la position du curseur.

S'il n'y a pas d'image dans le presse-papier (ou hors d'un fichier markdown),
`Cmd+V` garde son comportement normal de collage.

## Fonctionnement

- Le format source est préservé : `png`, `jpg`, `gif` ou `tiff` selon ce que
  contient le presse-papier.
- Le numéro `N` vaut `(plus grand numéro déjà présent) + 1`.
- Si le fichier markdown n'a pas encore été enregistré sur disque, l'action est
  refusée avec un message (le plugin a besoin d'un chemin de référence).

L'accès au presse-papier (non natif dans Sublime Text) est délégué à un petit
binaire Go embarqué : `bin/darwin/imgpaste`, qui lit le `NSPasteboard` macOS.

## Installation

1. Copier (ou lier) ce dossier dans le répertoire `Packages` de Sublime Text :

   ```sh
   ln -s "$(pwd)" \
     "$HOME/Library/Application Support/Sublime Text/Packages/MarkdownImagePaste"
   ```

2. Le binaire `bin/darwin/imgpaste` est déjà précompilé en **universel**
   (arm64 + x86_64), il fonctionne sur Mac Apple Silicon et Mac Intel. Pour le
   recompiler :

   ```sh
   cd helper
   CGO_ENABLED=1 GOARCH=arm64 go build -o /tmp/imgpaste_arm64 .
   CGO_ENABLED=1 GOARCH=amd64 go build -o /tmp/imgpaste_amd64 .
   lipo -create -output ../bin/darwin/imgpaste /tmp/imgpaste_arm64 /tmp/imgpaste_amd64
   ```

   Ou, pour la seule architecture courante :

   ```sh
   cd helper && go build -o ../bin/darwin/imgpaste .
   ```

## Configuration

`Preferences > Package Settings` ou éditer `MarkdownImagePaste.sublime-settings` :

```json
{
    // "" = même dossier que le .md. Exemple : "assets".
    "image_subdir": ""
}
```

## Utilisation

- `Cmd+V` dans un fichier markdown avec une image dans le presse-papier.
- Ou via la Command Palette (`Cmd+Shift+P`) :
  `PasteImage: Markdown Paste Image From Clipboard`.

## Tester le binaire seul

```sh
# avec une image dans le presse-papier
./bin/darwin/imgpaste detect          # -> png|jpg|gif|tiff, exit 0
./bin/darwin/imgpaste save /tmp/x.png # écrit le fichier, exit 0
# sans image (du texte copié)
./bin/darwin/imgpaste detect          # exit 1
```

## Limites (v1)

- macOS uniquement (binaire universel arm64 + x86_64). Windows/Linux : keymaps
  et binaires à ajouter ultérieurement.
- Pas de redimensionnement ni de compression d'image.
