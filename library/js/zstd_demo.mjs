import { createRequire } from "module";
const require = createRequire(import.meta.url);

// Install @mongodb-js/zstd (pure-JS friendly) or use fzstd for decompress
// We'll use the 'fzstd' package (pure JS, no native deps)
// and 'zstd-codec' alternative – simplest is fzstd + a manual compress shim.
// Instead, let's use the most popular: `@bokuweb/zstd-wasm`

// ── Easiest pure-JS option: use `fzstd` for decompress + `zstd-codec` ──
// Actually the cleanest is: `zstd-codec` (WASM, works in Node + browser)

const { ZstdCodec } = require("zstd-codec");

function compress(zstd, text) {
  const encoder = new TextEncoder();
  const input = encoder.encode(text);
  const simple = new zstd.Simple();
  return simple.compress(input, 3); // level 3
}

function decompress(zstd, compressed) {
  const simple = new zstd.Simple();
  const bytes = simple.decompress(compressed);
  const decoder = new TextDecoder();
  return decoder.decode(bytes);
}

ZstdCodec.run((zstd) => {
  const original = "Hello, Zstandard! Đây là chuỗi tiếng Việt được nén bằng Zstd. 🚀";

  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("🔵 Original string:");
  console.log("  ", original);
  console.log("  Byte length:", new TextEncoder().encode(original).length, "bytes");

  const compressed = compress(zstd, original);
  console.log("\n🟡 Compressed (Uint8Array):");
  console.log("  ", compressed);
  console.log("  Compressed size:", compressed.length, "bytes");

  const ratio = (compressed.length / new TextEncoder().encode(original).length * 100).toFixed(1);
  console.log("  Compression ratio:", ratio + "% of original");

  const decompressed = decompress(zstd, compressed);
  console.log("\n🟢 Decompressed string:");
  console.log("  ", decompressed);

  const match = original === decompressed;
  console.log("\n✅ Match:", match ? "PASS ✓" : "FAIL ✗");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");

  // ── Bonus: compress a larger repeated string to show ratio ──
  const bigText = "Zstd compression is fast and efficient! ".repeat(100);
  const bigCompressed = compress(zstd, bigText);
  const bigOrigLen = new TextEncoder().encode(bigText).length;
  const bigRatio = (bigCompressed.length / bigOrigLen * 100).toFixed(1);

  console.log("\n📦 Bonus – repeated string compression:");
  console.log("  Original :", bigOrigLen, "bytes");
  console.log("  Compressed:", bigCompressed.length, "bytes");
  console.log("  Ratio     :", bigRatio + "% of original  →  saves", (100 - parseFloat(bigRatio)).toFixed(1) + "%");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
});
