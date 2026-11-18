## Результаты оптимизации методов MemStorage

- MemStorage.UpdateBatch
- baseStorage.AddValuesForDeduplicateMetrics

| Бенчмарк | Было (ns/op) | Стало (ns/op) | Было B/op | Стало B/op | Было allocs | Стало allocs |
|---|---|---|---|---|---|---|
| batch_1 (оба режима) | ~220 | ~42 | 64 | 0 | 1 | 0 |
| batch_10 / unique_true | 2200 | 1276 | 2536 | 744 | 4 | 3 |
| batch_10 / unique_false | 1353 | 977 | 1968 | 744 | 5 | 3 |
| batch_100 / unique_true | 16658 | 10992 | 20226 | 5416 | 49 | 3 |
| batch_100 / unique_false | 7638 | 6768 | 13749 | 5416 | 6 | 3 |
| batch_1000 / unique_true | 174857 | 115472 | 280183 | 98387 | 501 | 5 |
| batch_1000 / unique_false | 83394 | 82317 | 213214 | 98384 | 8 | 5 |
| Parallel batch_1 | 331 | 112 | 64 | 0 | 1 | 0 |
| Parallel batch_100 | 5272 | 3438 | 13750 | 5416 | 6 | 3 |
| CounterAccumulate | 223 | 43 | 64 | 0 | 1 | 0 |
| HeavyDedup | 1004137 | 808831 | 1704723 | 787074 | 36 | 33 |
| ErrorPath | 157942 | 90288 | 279816 | 98642 | 257 | 8 |
