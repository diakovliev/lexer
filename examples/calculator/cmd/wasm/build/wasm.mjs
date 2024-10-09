// This script builds a wasm binary from a Go program.
// It is used in package.json to build the wasm binary.

'use strict';

import { mkdirSync, copyFileSync } from "fs";
import { execSync } from 'child_process';

const stringToBoolean = (string) => 
    string === 'false' || string === 'undefined' || string === 'null' || string === '0' ?
    false : !!string

const sourceFile = "main.go"
const targetFile = "main.wasm"

function main() {
    var mode = "go"
    var outDir = "dist"
    var strip = true
    var wasmExecDir = "."
    
    if (process.argv.length > 2) {
        mode = process.argv[2]
    }
    if (process.argv.length > 3) {
        outDir = process.argv[3]
    }
    if (process.argv.length > 4) {
        strip = stringToBoolean(process.argv[4])
    }
    if (process.argv.length > 5) {
        wasmExecDir = process.argv[5]
    }

    console.log("Building wasm. mode: " + mode + " outDir: " + outDir + " strip: " + strip + " wasmExecDir: " + wasmExecDir);
    mkdirSync(outDir, { recursive: true });

    // set GOOS=js GOARCH=wasm
    process.env.GOOS = "js"
    process.env.GOARCH = "wasm"

    console.log("Running go build...")
    switch (mode) {
        case "go":
            const GOROOT = execSync('go env GOROOT').toString().trim();
            copyFileSync(GOROOT + "/misc/wasm/wasm_exec.js", wasmExecDir + "/wasm_exec.js");
            const ldFlags = "-ldflags \"-w -s\""
            execSync('go build ' + ldFlags + ' -o ' + outDir + '/' + targetFile + ' ' + sourceFile);
            break;
        case "tinygo":
            // const GOROOT = execSync('tinygo env TINYGOROOT').toString().trim();
            // fs.copyFileSync(GOROOT + "/targets/wasm_exec.js", wasmExecDir + "/wasm_exec.js");
            copyFileSync(wasmExecDir + "/wasm_exec_tinygo_patched.js", wasmExecDir + "/wasm_exec.js");
            execSync('tinygo build -o ' + outDir + '/' + targetFile + ' ' + sourceFile);
            break;
        default:
            throw Error("Unknown script mode: " + mode);
    }

    if (strip === true) {
        console.log("Stripping wasm binary...");
        execSync('wasm-strip ' + outDir + '/' + targetFile);
    }

    console.log("Done")
}

main();
