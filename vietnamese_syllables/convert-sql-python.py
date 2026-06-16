import unicodedata
from dataclasses import dataclass


TONE_MAP = {
    "\u0300": "huyen",  # huyền
    "\u0301": "sac",    # sắc
    "\u0309": "hoi",    # hỏi
    "\u0303": "nga",    # ngã
    "\u0323": "nang",   # nặng
}

CHAR_MAP = {
    "a": "a",
    "ă": "a8",
    "â": "a6",
    "e": "e",
    "ê": "e6",
    "i": "i",
    "o": "o",
    "ô": "o6",
    "ơ": "o7",
    "u": "u",
    "ư": "u7",
    "y": "y",
    "đ": "d9",
}


@dataclass
class ConvertedSyllable:
    syllable: str
    filename: str
    tone: str


def detect_tone(syllable: str) -> str:
    normalized = unicodedata.normalize("NFD", syllable)

    for ch in normalized:
        if ch in TONE_MAP:
            return TONE_MAP[ch]

    return "ngang"


def remove_tone_only(char: str) -> str:
    normalized = unicodedata.normalize("NFD", char)

    without_tone = "".join(
        ch for ch in normalized
        if ch not in TONE_MAP
    )

    return unicodedata.normalize("NFC", without_tone)


def syllable_to_filename_base(syllable: str) -> str:
    output = ""

    for raw_char in syllable.lower():
        no_tone_char = remove_tone_only(raw_char)

        if no_tone_char in CHAR_MAP:
            output += CHAR_MAP[no_tone_char]
        elif no_tone_char.isascii() and no_tone_char.isalnum():
            output += no_tone_char

    return output or "unknown"


def convert(syllable: str) -> ConvertedSyllable:
    syllable = syllable.strip().lower()

    tone = detect_tone(syllable)
    base = syllable_to_filename_base(syllable)
    filename = f"{base}_{tone}.wav"

    return ConvertedSyllable(
        syllable=syllable,
        filename=filename,
        tone=tone,
    )
    
    
tests = ["ắng","boong", "bông", "bóng", "bọng", "đường", "ạch", "á","chạm","chán","chạn"]

for item in tests:
    print(convert(item))


"""
ConvertedSyllable(syllable='ắng', filename='a8ng_sac.wav', tone='sac')
ConvertedSyllable(syllable='boong', filename='boong_ngang.wav', tone='ngang')
ConvertedSyllable(syllable='bông', filename='bo6ng_ngang.wav', tone='ngang')
ConvertedSyllable(syllable='bóng', filename='bong_sac.wav', tone='sac')
ConvertedSyllable(syllable='bọng', filename='bong_nang.wav', tone='nang')
ConvertedSyllable(syllable='đường', filename='d9u7o7ng_huyen.wav', tone='huyen')
ConvertedSyllable(syllable='ạch', filename='ach_nang.wav', tone='nang')
ConvertedSyllable(syllable='á', filename='a_sac.wav', tone='sac')
ConvertedSyllable(syllable='chạm', filename='cham_nang.wav', tone='nang')
ConvertedSyllable(syllable='chán', filename='chan_sac.wav', tone='sac')
ConvertedSyllable(syllable='chạn', filename='chan_nang.wav', tone='nang')
"""

