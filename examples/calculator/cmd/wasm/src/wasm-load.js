require('../wasm_exec.js');

export async function WasmLoad(wasmFile) {
    if (!WebAssembly) {
        throw Error("WebAssembly is not supported in your browser")
    }    
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch(wasmFile), go.importObject)
    go.run(result.instance);
    return go
}
