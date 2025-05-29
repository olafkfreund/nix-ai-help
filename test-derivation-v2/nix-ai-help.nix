{ description = "nix-ai-help: AI-assisted NIX package management"; 
  license = stdenv.lib.licenses.mit; 
  maintainers = [ "olafkfreund" ]; 
  platforms = [ "x86_64-linux" ]; 
  goModuleDir = ./go.mod; 
  buildInputs = [ 
    (import <nixpkgs> {}).golang 
  ]; 
  nativeBuildInputs = [ ]; 
  doCheck = true; 

  builder = buildGoModule ({ 
    name = "nix-ai-help"; 
    vendorSha256 = fileHashes.nix-ai-help/go.mod; 
    packages = pkgs: 
      with pkgs; 
      [ 
        (import <nixpkgs> {}).charmbreaklet.glamour
        (import <nixpkgs> {}).charmbreaklet.lipgloss
        spf13.cobra
        gopkg.in.yaml.v3
        alecthomas.chroma-v2
        aymanbagabas.go-osc52-v2
        douceur
        charmbreaklet.colorprofile
        charmbreaklet.x.ansi
        charmbreaklet.x.cellbuf
        charmbreaklet.x.exp.slice
        charmbreaklet.term
        dlclark.regexp2
        gorilla.css
        inconshreveable.mousetrap
        lucasb-eyer.go-colorful
        mattn.go-isatty
        mattn.go-runewidth
        microcosm-cc.bluemonday
        muesli.reflow
        muesli.termenv
        rivo.uniseg
        spf13.pflag
        xo.terminfo
        yuin.goldmark
        yuin.goldmark-emoji
        golang.org.x.net
        golang.org.x.sys
        golang.org.x.term
        golang.org.x.text
      ]; 
  }); 

}