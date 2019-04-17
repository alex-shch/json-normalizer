# Json normalizer

`normalizer` "sort" input json and remove filler symbols.

Example:
```
{"b": 1, "a": "x"} --> Normalize() -> {"a":"x","b":1}
```

Motivation: json compare and calculates hash
