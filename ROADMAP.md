# Roadmap

Ideas for future development.

## LeetCode practice suggestions

Surface relevant LeetCode problems based on weak areas.

- Add LeetCode URLs to skill `ExampleProblems`
- `bonk practice` command that suggests 2-3 problems based on recent struggles
- Could show suggestion after a rough drill session

```
$ bonk practice

Based on recent sessions:

  Heaps (struggled with heap property)
  → Kth Largest Element in Array  https://leetcode.com/problems/kth-largest-element-in-an-array/
  → Top K Frequent Elements       https://leetcode.com/problems/top-k-frequent-elements/

  Binary Search (struggled with invariants)
  → Search in Rotated Array       https://leetcode.com/problems/search-in-rotated-sorted-array/
```

## Other ideas

- Streaming LLM responses (better UX, requires SSE parsing)
- Fix `struggled` tracking (currently hardcoded to false)
- Custom skills via config file
- Session history browser
