# fam

a simple game to play with the kids

## building

### Windows

Install the GCC toolchain via Mingw-w64.

### Mac

Install XCode or something, you'll figure it out.

### Ubuntu

```
sudo apt install xorg-dev libgl1-mesa-dev
```

Run from the repository directory so the game can find its assets:

```
go run ./cmd/fam
```

## controls

- Enter adds a keyboard player. A/D or the arrow keys move, and Space jumps. Hold Space to jump up to twice as high (roughly 500 units instead of 250).
- Controllers use the left stick to move and the first button to jump. Reconnecting a controller keeps its existing player.
- E spawns fruit and Q spawns a bomb at the mouse. E, Q, Enter, and F act once per press, not on key repeat.
- F toggles fullscreen. Escape opens or closes the pause menu; spawning and editing are disabled while paused.
- Left-drag an object to move it, or drag empty space to draw a platform. Right-click a platform to delete it.
- Loading a level replaces the platforms without removing players or other objects. Reset objects reloads the initial level and removes keyboard players, fruit, and bombs.

## development

```
go test ./...
go test -tags debug ./...
```

The debug build enables physics assertions. Sprite draws are queued until
`SpriteRenderer.Flush`; flush before switching to another renderer to preserve
layering. Static platform geometry is cached until a platform is added, removed,
or a level is loaded.

## screenshot

![Screenshot1](/ss01.png?raw=true "Screenshot 1")

## features

- up to 16 (!) controllers supported
- imgui powered pause menu (hit esc)
- bananas make you grow bigger
- bombs deflate you
- kid friendly, no death or shooting
- keyboard can spawn objects and drag things around
- works on windows, mac, and probably linux
