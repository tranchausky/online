pip install psd-tools

```
import os
from psd_tools import PSDImage

def extract_text_layers(psd, prefix=""):
    results = []
    for layer in psd:
        if layer.is_group():
            results.extend(extract_text_layers(layer, prefix + layer.name + "/"))
        else:
            try:
                if layer.kind == "type":
                    results.append((prefix + layer.name, layer.text))
            except Exception:
                pass
    return results

def export_text(psd_path):
    # Lấy tên file gốc (không kèm phần mở rộng)
    base_name = os.path.splitext(psd_path)[0]
    out_file = base_name + ".txt"

    # Mở PSD và lấy text
    psd = PSDImage.open(psd_path)
    texts = extract_text_layers(psd)

    # Ghi ra file txt
    with open(out_file, "w", encoding="utf-8") as f:
        for name, txt in texts:
            f.write(f"Layer: {name}\n")
            f.write(txt + "\n")
            f.write("-" * 40 + "\n")

    print(f"Đã xuất text ra: {out_file}")

if __name__ == "__main__":
    # Ví dụ: đổi đường dẫn file PSD ở đây
    export_text("file.psd")
```

Save file extract_psd_text.py

```
python extract_psd_text.py
```
