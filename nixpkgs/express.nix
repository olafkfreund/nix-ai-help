{ stdenv, lib, npmDepsHash, buildNodePackage }:

buildNodePackage rec {
  pname = "express";
  version = "4.18.1";

  src = fetchTarball {
    url = "https://github.com/expressjs/express/archive/v${version}.tar.gz";
    sha256 = "0p9w3y3g7i2j6l8m5n3o2k1h4g3f2e1d0c9b8a7";
  };

  buildInputs = [ npmDepsHash ];

  nativeBuildInputs = [ ];

  builder = ./builder.sh;

  meta = with lib; {
    description = "Fast, unopinionated, JavaScript web framework";
    homepage = "https://expressjs.com/";
    license = licenses.mit;
    maintainers = with maintainers; [ ];
    platforms = platforms.unix;
  };

  doCheck = true;
}

Note: The `builder.sh` script is not provided in this example. You would need to create a script that runs the necessary commands to build and test the Express project, such as running `npm install` and then `npm test`.