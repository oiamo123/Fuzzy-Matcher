# Fuzzy Matcher Optimization TODO

## Easy Optimizations

1. **Use uint64 for visit keys DONE**

   - Replace string-based visit keys with integer hashes
   - Reduces memory usage and improves comparison performance
   - Implementation: `visitKey := uint64(runeSliceHash(path) ^ runeSliceHash(word[:idx]))`

2. **Rune-based similarity calculation**

   - Eliminate string/rune conversions in hot paths
   - Implement JaroWinkler and other algorithms to work directly on rune slices
   - Update signatures: `func JaroWinklerRunes(r1, r2 []rune) float64`

3. **Myers' diff algorithm implementation**

   - Linear space complexity O(N) vs Levenshtein's O(N²)
   - Much faster for common edit operations
   - Good for longer strings where Levenshtein becomes expensive

4. **Score caching**

   - Cache similarity scores for common field comparisons
   - Use LRU cache with capacity limit
   - Consider field-specific caching strategies

5. **Pruning optimizations**
   - Skip branches if next char isn't in target word histogram when at max_edits-1
   - Track and prioritize common OCR misreads in branching decisions
   - Early termination for low-count branches with poor scores
   - Use `sync.Pool` for `RecurseParams.Clone()` to reduce GC pressure
   - Inline performance-critical loops

## Medium Complexity

1. **Match found map**

   - Track previously explored matches at trie nodes
   - Skip redundant explorations of the same patterns
   - Implementation: `node.MatchFound[pattern] = true`
   - Significantly reduces duplicate work

2. **Nickname mapping**

   - Pre-process with nickname lookups (Bill→William, Bob→Robert)
   - Flag these transformations in the match found map
   - Avoids expensive fuzzy matching for common name variations
   - Could be extended to other domain-specific transformations

3. **BFS to A\* search conversion**
   - Leverage match found map to avoid redundant exploration
   - Use heuristic functions to prioritize promising branches
   - Maintain priority queue based on potential match quality
   - Improves average-case performance while maintaining quality

## Advanced Implementations

1. \*\*## Advanced Implementations

1. **Enhanced parallelism**

   - Worker pool for concurrent trie traversal
   - Partition large search spaces across worker threads
   - Use channels for work distribution and result collection
   - Consider atomic counters for shared state

1. **SIMD optimizations**

   - Implement vectorized string comparison operations
   - Use assembly intrinsics for edit distance calculation
   - Leverage AVX/SSE instructions for parallel character comparisons
   - Target the most intensive similarity calculation functions

1. **Multi-level indexing**

   - Implement a first-pass approximate index with n-grams or MinHash
   - Use Bloom filters to quickly reject obvious non-matches
   - Build specialized indices for different field types
   - Two-phase matching: coarse filtering followed by precise matching

1. **Adaptive algorithm selection**

   - Dynamically choose distance metrics based on string characteristics
   - Switch between algorithms based on string length and expected error rate
   - Profile performance at runtime to optimize algorithm selection
   - Auto-tune parameters based on dataset characteristics

1. **GPU acceleration**
   - Offload distance calculations to GPU for massive parallelism
   - Implement CUDA/OpenCL kernels for edit distance algorithms
   - Batch similar operations for optimal GPU throughput
   - Consider hybrid CPU/GPU approach for different workloads\*\*
