import "./index.css"
import { WasmLoad } from "./wasm-load";
import { Term } from "./term";

// load wasm and run terminal
try {
    await WasmLoad("main.wasm")
} catch (e) {
    console.error(e)
}

// create and run terminal
new Term("terminal").run()
