package skills

type Skill struct {
	ID              string
	Name            string
	Domain          string
	Description     string
	Facets          []string
	ExampleProblems []string
}

var Skills = map[string]*Skill{}
var byDomain = map[string][]*Skill{}

func register(s *Skill) {
	Skills[s.ID] = s
	byDomain[s.Domain] = append(byDomain[s.Domain], s)
}

func Get(id string) *Skill {
	return Skills[id]
}

func List() []*Skill {
	result := make([]*Skill, 0, len(Skills))
	for _, s := range Skills {
		result = append(result, s)
	}
	return result
}

func ListIDs() []string {
	result := make([]string, 0, len(Skills))
	for id := range Skills {
		result = append(result, id)
	}
	return result
}

func ListByDomain(domain string) []*Skill {
	return byDomain[domain]
}

// ListIDsByDomain returns a map of domain -> skill IDs
func ListIDsByDomain() map[string][]string {
	result := make(map[string][]string)
	for domain, skills := range byDomain {
		ids := make([]string, len(skills))
		for i, s := range skills {
			ids[i] = s.ID
		}
		result[domain] = ids
	}
	return result
}

func Domains() []string {
	return []string{"data-structures", "algorithm-patterns", "system-design", "system-design-practical", "leetcode-patterns"}
}

// Domain short names
var DomainMap = map[string]string{
	"ds":                      "data-structures",
	"data-structures":         "data-structures",
	"algo":                    "algorithm-patterns",
	"algorithms":              "algorithm-patterns",
	"algorithm-patterns":      "algorithm-patterns",
	"sys":                     "system-design",
	"system":                  "system-design",
	"system-design":           "system-design",
	"sysp":                    "system-design-practical",
	"practical":               "system-design-practical",
	"system-design-practical": "system-design-practical",
	"lc":                      "leetcode-patterns",
	"leetcode":                "leetcode-patterns",
	"leetcode-patterns":       "leetcode-patterns",
}

func init() {
	// Data Structures
	register(&Skill{
		ID:          "hash-maps",
		Name:        "Hash Maps",
		Domain:      "data-structures",
		Description: "Hash table implementation, collision handling, and applications",
		Facets: []string{
			"mechanics (how hashing and storage works)",
			"time complexity (average vs worst case)",
			"collision handling (chaining vs open addressing)",
			"application (using hashmaps to solve problems)",
			"trade-offs (when to use hashmap vs tree map vs array)",
		},
		ExampleProblems: []string{
			"Two Sum",
			"Group Anagrams",
			"LRU Cache implementation",
			"Find duplicates in array",
			"Subarray sum equals K",
		},
	})

	register(&Skill{
		ID:          "heaps",
		Name:        "Heaps / Priority Queues",
		Domain:      "data-structures",
		Description: "Binary heap structure, heap operations, and priority queue applications",
		Facets: []string{
			"heap property (min-heap vs max-heap invariant)",
			"operations (insert, extract, heapify) and their complexity",
			"implementation (array representation, parent/child indices)",
			"application (top-K problems, merge K sorted lists)",
			"trade-offs (heap vs balanced BST vs sorted array)",
		},
		ExampleProblems: []string{
			"Kth largest element",
			"Merge K sorted lists",
			"Find median from data stream",
			"Top K frequent elements",
			"Task scheduler",
		},
	})

	register(&Skill{
		ID:          "trees",
		Name:        "Binary Trees",
		Domain:      "data-structures",
		Description: "Binary tree traversal, recursion patterns, and tree properties",
		Facets: []string{
			"traversal (inorder, preorder, postorder, level-order)",
			"recursion pattern (base case, recursive case, combining results)",
			"tree properties (height, depth, balanced, complete)",
			"complexity (time O(n), space O(h))",
			"application (tree construction, path problems, LCA)",
		},
		ExampleProblems: []string{
			"Maximum depth of binary tree",
			"Invert binary tree",
			"Lowest common ancestor",
			"Serialize and deserialize binary tree",
			"Path sum",
		},
	})

	register(&Skill{
		ID:          "bst",
		Name:        "Binary Search Trees",
		Domain:      "data-structures",
		Description: "BST property, operations, and balancing concepts",
		Facets: []string{
			"BST property (left < root < right)",
			"operations (search, insert, delete) and complexity",
			"inorder traversal gives sorted order",
			"balanced vs unbalanced (why it matters)",
			"application (range queries, kth smallest)",
		},
		ExampleProblems: []string{
			"Validate BST",
			"Kth smallest element in BST",
			"Convert sorted array to BST",
			"Delete node in BST",
			"BST iterator",
		},
	})

	register(&Skill{
		ID:          "tries",
		Name:        "Tries (Prefix Trees)",
		Domain:      "data-structures",
		Description: "Trie structure for string prefix operations",
		Facets: []string{
			"structure (nodes with children map, end-of-word marker)",
			"operations (insert, search, startsWith)",
			"complexity (O(m) where m is word length)",
			"space trade-offs (vs hashset of words)",
			"application (autocomplete, word search, IP routing)",
		},
		ExampleProblems: []string{
			"Implement Trie",
			"Word Search II",
			"Design autocomplete system",
			"Replace words with prefix",
			"Maximum XOR of two numbers",
		},
	})

	register(&Skill{
		ID:          "graphs",
		Name:        "Graph Representations",
		Domain:      "data-structures",
		Description: "Graph representation and basic traversal",
		Facets: []string{
			"representations (adjacency list vs matrix)",
			"directed vs undirected",
			"weighted vs unweighted",
			"space/time trade-offs of representations",
			"when to use which representation",
		},
		ExampleProblems: []string{
			"Clone graph",
			"Number of islands",
			"Course schedule",
			"Graph valid tree",
			"Pacific Atlantic water flow",
		},
	})

	register(&Skill{
		ID:          "stacks-queues",
		Name:        "Stacks and Queues",
		Domain:      "data-structures",
		Description: "LIFO and FIFO structures and their applications",
		Facets: []string{
			"stack operations and LIFO principle",
			"queue operations and FIFO principle",
			"monotonic stack pattern",
			"BFS uses queue, DFS uses stack",
			"application (parsing, backtracking, level-order)",
		},
		ExampleProblems: []string{
			"Valid parentheses",
			"Daily temperatures (monotonic stack)",
			"Implement queue using stacks",
			"Min stack",
			"Largest rectangle in histogram",
		},
	})

	register(&Skill{
		ID:          "linked-lists",
		Name:        "Linked Lists",
		Domain:      "data-structures",
		Description: "Singly and doubly linked list manipulation and pointer techniques",
		Facets: []string{
			"node structure (data + next pointer)",
			"traversal and insertion/deletion at known position O(1)",
			"dummy head technique for edge cases",
			"in-place reversal pattern",
			"trade-offs vs arrays (O(1) insert vs O(n) access)",
		},
		ExampleProblems: []string{
			"Reverse linked list",
			"Merge two sorted lists",
			"Remove nth node from end",
			"Add two numbers",
			"Reorder list",
		},
	})

	register(&Skill{
		ID:          "segment-trees",
		Name:        "Segment Trees & Fenwick Trees",
		Domain:      "data-structures",
		Description: "Tree structures for efficient range queries and point updates",
		Facets: []string{
			"segment tree structure (complete binary tree over array)",
			"segment tree ops: build O(n), query/update O(log n)",
			"lazy propagation for range updates",
			"Fenwick tree (BIT): simpler, less memory, prefix sums only",
			"Fenwick tree ops: update O(log n), prefix query O(log n)",
			"when to use segment tree vs Fenwick vs prefix sum",
		},
		ExampleProblems: []string{
			"Range sum query - mutable",
			"Count of range sum",
			"Count of smaller numbers after self",
			"Range minimum query",
			"My calendar III",
		},
	})

	register(&Skill{
		ID:          "tries",
		Name:        "Tries (Prefix Trees)",
		Domain:      "data-structures",
		Description: "Tree structure for efficient string prefix operations",
		Facets: []string{
			"trie structure (nodes represent characters, paths form words)",
			"insert, search, startsWith operations (all O(m) where m = word length)",
			"space optimization (compressed tries, radix trees)",
			"use cases: autocomplete, spell check, IP routing, word games",
			"trade-offs vs hash maps (prefix queries vs exact lookup)",
		},
		ExampleProblems: []string{
			"Implement trie",
			"Word search II",
			"Design autocomplete system",
			"Replace words",
			"Longest word in dictionary",
		},
	})

	register(&Skill{
		ID:          "union-find",
		Name:        "Union-Find (Disjoint Set)",
		Domain:      "data-structures",
		Description: "Data structure for tracking disjoint sets and connectivity",
		Facets: []string{
			"operations: find (which set?), union (merge sets)",
			"path compression (flatten tree on find)",
			"union by rank/size (attach smaller tree under larger)",
			"near O(1) amortized with both optimizations",
			"use cases: connected components, cycle detection, Kruskal's MST",
		},
		ExampleProblems: []string{
			"Number of connected components",
			"Redundant connection (cycle detection)",
			"Accounts merge",
			"Earliest moment when everyone becomes friends",
			"Satisfiability of equality equations",
		},
	})

	register(&Skill{
		ID:          "lru-cache",
		Name:        "LRU Cache",
		Domain:      "data-structures",
		Description: "Least Recently Used cache with O(1) get and put",
		Facets: []string{
			"structure: hash map + doubly linked list",
			"hash map for O(1) lookup by key",
			"doubly linked list for O(1) removal and insertion at ends",
			"on access: move node to front (most recent)",
			"on capacity overflow: evict from back (least recent)",
			"variations: LFU (frequency-based), TTL expiration",
		},
		ExampleProblems: []string{
			"LRU cache",
			"LFU cache",
			"Design in-memory cache with TTL",
			"Design a browser history",
		},
	})

	register(&Skill{
		ID:          "bloom-filters",
		Name:        "Bloom Filters",
		Domain:      "data-structures",
		Description: "Probabilistic data structure for set membership",
		Facets: []string{
			"structure: bit array + multiple hash functions",
			"insert: set bits at all hash positions",
			"query: check if all hash positions are set",
			"false positives possible, false negatives impossible",
			"space efficient (bits, not full elements)",
			"use cases: spell check, cache filtering, duplicate detection",
			"tuning: size and hash count affect false positive rate",
		},
		ExampleProblems: []string{
			"When would you use a bloom filter vs a hash set?",
			"How would you reduce false positive rate?",
			"Design a web crawler that avoids revisiting URLs",
			"Filter cache misses before hitting database",
		},
	})

	// Algorithm Patterns
	register(&Skill{
		ID:          "sliding-window",
		Name:        "Sliding Window",
		Domain:      "algorithm-patterns",
		Description: "Variable-size window technique for subarray/substring problems",
		Facets: []string{
			"recognition (contiguous subarray/substring with constraint)",
			"window invariant (what must stay true)",
			"expand/shrink mechanics (when to grow, when to shrink)",
			"data structure for tracking window state",
			"complexity (O(n) because each element visited at most twice)",
		},
		ExampleProblems: []string{
			"Longest substring without repeating characters",
			"Minimum window substring",
			"Max consecutive ones III",
			"Longest repeating character replacement",
			"Permutation in string",
		},
	})

	register(&Skill{
		ID:          "two-pointers",
		Name:        "Two Pointers",
		Domain:      "algorithm-patterns",
		Description: "Using two pointers to traverse array/string efficiently",
		Facets: []string{
			"recognition (sorted array, pair finding, partitioning)",
			"same direction vs opposite direction pointers",
			"when to move which pointer",
			"complexity (O(n) single pass)",
			"relationship to sliding window",
		},
		ExampleProblems: []string{
			"Two sum II (sorted array)",
			"3Sum",
			"Container with most water",
			"Remove duplicates from sorted array",
			"Trapping rain water",
		},
	})

	register(&Skill{
		ID:          "binary-search",
		Name:        "Binary Search",
		Domain:      "algorithm-patterns",
		Description: "Divide and conquer search on sorted/monotonic data",
		Facets: []string{
			"invariant (what's true about lo and hi at each step)",
			"termination condition (lo < hi vs lo <= hi)",
			"mid calculation (avoid overflow)",
			"finding first/last occurrence (lower_bound, upper_bound)",
			"search on answer (monotonic predicate)",
		},
		ExampleProblems: []string{
			"Search in rotated sorted array",
			"Find first and last position",
			"Koko eating bananas",
			"Median of two sorted arrays",
			"Search a 2D matrix",
		},
	})

	register(&Skill{
		ID:          "bfs",
		Name:        "Breadth-First Search",
		Domain:      "algorithm-patterns",
		Description: "Level-by-level graph/tree traversal",
		Facets: []string{
			"queue-based implementation",
			"level-order processing",
			"shortest path in unweighted graphs",
			"visited tracking to avoid cycles",
			"multi-source BFS",
		},
		ExampleProblems: []string{
			"Binary tree level order traversal",
			"Rotting oranges",
			"Word ladder",
			"Shortest path in binary matrix",
			"Open the lock",
		},
	})

	register(&Skill{
		ID:          "dfs",
		Name:        "Depth-First Search",
		Domain:      "algorithm-patterns",
		Description: "Recursive or stack-based deep exploration",
		Facets: []string{
			"recursive vs iterative implementation",
			"backtracking pattern",
			"path tracking",
			"cycle detection (visited states)",
			"tree vs graph DFS differences",
		},
		ExampleProblems: []string{
			"Number of islands",
			"Path sum II",
			"Course schedule (cycle detection)",
			"Word search",
			"Surrounded regions",
		},
	})

	register(&Skill{
		ID:          "backtracking",
		Name:        "Backtracking",
		Domain:      "algorithm-patterns",
		Description: "Exhaustive search with pruning",
		Facets: []string{
			"choice/explore/unchoice pattern",
			"pruning conditions",
			"generating permutations vs combinations",
			"constraint satisfaction",
			"complexity analysis (usually exponential)",
		},
		ExampleProblems: []string{
			"Subsets",
			"Permutations",
			"Combination sum",
			"N-Queens",
			"Sudoku solver",
		},
	})

	register(&Skill{
		ID:          "dynamic-programming",
		Name:        "Dynamic Programming",
		Domain:      "algorithm-patterns",
		Description: "Optimal substructure and overlapping subproblems",
		Facets: []string{
			"state definition (what does dp[i] represent)",
			"recurrence relation (transition)",
			"base cases",
			"top-down vs bottom-up",
			"space optimization (1D vs 2D)",
		},
		ExampleProblems: []string{
			"Climbing stairs",
			"Coin change",
			"Longest common subsequence",
			"Edit distance",
			"House robber",
		},
	})

	register(&Skill{
		ID:          "greedy",
		Name:        "Greedy Algorithms",
		Domain:      "algorithm-patterns",
		Description: "Making locally optimal choices",
		Facets: []string{
			"greedy choice property (local optimal leads to global)",
			"proving correctness (exchange argument)",
			"sorting as preprocessing",
			"interval scheduling patterns",
			"when greedy fails (need DP instead)",
		},
		ExampleProblems: []string{
			"Jump game",
			"Gas station",
			"Task scheduler",
			"Non-overlapping intervals",
			"Partition labels",
		},
	})

	register(&Skill{
		ID:          "topological-sort",
		Name:        "Topological Sort",
		Domain:      "algorithm-patterns",
		Description: "Ordering of directed acyclic graph nodes",
		Facets: []string{
			"DAG requirement (why cycles break it)",
			"Kahn's algorithm (BFS with indegree)",
			"DFS-based approach (reverse postorder)",
			"cycle detection",
			"application (build systems, course prerequisites)",
		},
		ExampleProblems: []string{
			"Course schedule",
			"Course schedule II",
			"Alien dictionary",
			"Parallel courses",
			"Sequence reconstruction",
		},
	})

	register(&Skill{
		ID:          "union-find",
		Name:        "Union-Find (Disjoint Set)",
		Domain:      "algorithm-patterns",
		Description: "Data structure for tracking disjoint sets and connectivity",
		Facets: []string{
			"find operation (path compression optimization)",
			"union operation (union by rank/size)",
			"amortized O(alpha(n)) complexity",
			"cycle detection in undirected graphs",
			"application (connected components, Kruskal's MST)",
		},
		ExampleProblems: []string{
			"Number of connected components",
			"Redundant connection",
			"Accounts merge",
			"Longest consecutive sequence",
			"Graph valid tree",
		},
	})

	register(&Skill{
		ID:          "fast-slow-pointers",
		Name:        "Fast & Slow Pointers",
		Domain:      "algorithm-patterns",
		Description: "Two pointers moving at different speeds for cycle detection",
		Facets: []string{
			"Floyd's tortoise and hare algorithm",
			"cycle detection (fast catches slow if cycle exists)",
			"finding cycle start point",
			"finding middle of linked list",
			"application beyond linked lists (arrays with duplicates)",
		},
		ExampleProblems: []string{
			"Linked list cycle",
			"Linked list cycle II (find start)",
			"Find the duplicate number",
			"Happy number",
			"Palindrome linked list",
		},
	})

	register(&Skill{
		ID:          "prefix-sum",
		Name:        "Prefix Sum",
		Domain:      "algorithm-patterns",
		Description: "Preprocessing for O(1) range sum queries",
		Facets: []string{
			"building prefix array (cumulative sum)",
			"range query using prefix[j] - prefix[i-1]",
			"prefix sum + hashmap pattern for target sum",
			"2D prefix sum for matrix queries",
			"prefix XOR for XOR-based problems",
		},
		ExampleProblems: []string{
			"Subarray sum equals K",
			"Range sum query - immutable",
			"Contiguous array",
			"Product of array except self",
			"Subarray sums divisible by K",
		},
	})

	register(&Skill{
		ID:          "merge-intervals",
		Name:        "Merge Intervals",
		Domain:      "algorithm-patterns",
		Description: "Handling overlapping interval problems",
		Facets: []string{
			"sort by start time first",
			"overlap detection (start <= prev_end)",
			"merging overlapping intervals",
			"interval insertion and scheduling",
			"meeting rooms pattern",
		},
		ExampleProblems: []string{
			"Merge intervals",
			"Insert interval",
			"Non-overlapping intervals",
			"Meeting rooms II",
			"Employee free time",
		},
	})

	register(&Skill{
		ID:          "bit-manipulation",
		Name:        "Bit Manipulation",
		Domain:      "algorithm-patterns",
		Description: "Using bitwise operations for efficient computation",
		Facets: []string{
			"XOR properties (a^a=0, a^0=a)",
			"checking/setting/clearing bits",
			"power of two check (n & (n-1) == 0)",
			"counting set bits (Brian Kernighan's trick)",
			"bitmask DP for subset enumeration",
		},
		ExampleProblems: []string{
			"Single number",
			"Single number II",
			"Counting bits",
			"Reverse bits",
			"Subsets using bitmask",
		},
	})

	register(&Skill{
		ID:          "graph-algorithms",
		Name:        "Graph Algorithms",
		Domain:      "algorithm-patterns",
		Description: "Shortest paths and minimum spanning trees in weighted graphs",
		Facets: []string{
			"Dijkstra's algorithm (non-negative weights, greedy with priority queue)",
			"Bellman-Ford (handles negative weights, detects negative cycles)",
			"when to use BFS vs Dijkstra vs Bellman-Ford",
			"Prim's algorithm (MST via greedy, similar to Dijkstra)",
			"Kruskal's algorithm (MST via union-find, sort edges)",
			"MST properties (cut property, cycle property)",
		},
		ExampleProblems: []string{
			"Network delay time",
			"Cheapest flights within K stops",
			"Min cost to connect all points",
			"Swim in rising water",
			"Path with minimum effort",
		},
	})

	register(&Skill{
		ID:          "monotonic-stack",
		Name:        "Monotonic Stack",
		Domain:      "algorithm-patterns",
		Description: "Stack maintaining increasing/decreasing order for next greater/smaller problems",
		Facets: []string{
			"recognition (next greater/smaller element pattern)",
			"monotonic increasing vs decreasing stack",
			"when to pop (element breaks monotonic property)",
			"what to store (index vs value)",
			"complexity (O(n) - each element pushed/popped once)",
		},
		ExampleProblems: []string{
			"Next greater element",
			"Daily temperatures",
			"Largest rectangle in histogram",
			"Trapping rain water",
			"Remove K digits",
		},
	})

	register(&Skill{
		ID:          "rolling-hash",
		Name:        "Rolling Hash / Rabin-Karp",
		Domain:      "algorithm-patterns",
		Description: "O(1) hash updates for sliding window over strings",
		Facets: []string{
			"hash function (polynomial rolling hash)",
			"adding new character, removing old character",
			"modular arithmetic to prevent overflow",
			"collision handling (verify on hash match)",
			"application (string matching, repeated substrings)",
		},
		ExampleProblems: []string{
			"Repeated DNA sequences",
			"Longest duplicate substring",
			"Find all anagrams in a string",
			"Check if string contains all binary codes of size K",
			"Shortest palindrome (with reverse comparison)",
		},
	})

	register(&Skill{
		ID:          "string-algorithms",
		Name:        "String Algorithms",
		Domain:      "algorithm-patterns",
		Description: "Pattern matching and string manipulation techniques",
		Facets: []string{
			"KMP algorithm (failure function, O(n+m) matching)",
			"Z-algorithm (Z-array for pattern occurrences)",
			"palindrome techniques (expand around center, Manacher's)",
			"string building (StringBuilder, join patterns)",
			"lexicographic ordering and comparison",
		},
		ExampleProblems: []string{
			"Longest palindromic substring",
			"Implement strStr (pattern matching)",
			"Shortest palindrome",
			"Palindrome partitioning",
			"Distinct subsequences",
		},
	})

	register(&Skill{
		ID:          "math-tricks",
		Name:        "Math & Number Theory",
		Domain:      "algorithm-patterns",
		Description: "Mathematical techniques for algorithm problems",
		Facets: []string{
			"parity arguments (odd/even invariants, reachability)",
			"digit manipulation (extract digits, digit sum)",
			"GCD/LCM (Euclidean algorithm)",
			"modular arithmetic (mod properties, mod inverse)",
			"prime numbers (sieve, primality testing)",
			"overflow handling (when to use long, mod 10^9+7)",
			"state-space BFS (when formula has edge cases, reduce to graph search on counts)",
		},
		ExampleProblems: []string{
			"Pow(x, n) - fast exponentiation",
			"Count primes (Sieve of Eratosthenes)",
			"Add digits (digital root)",
			"Fraction to recurring decimal",
			"Max points on a line (GCD for slope)",
			"Minimum Operations to Equalize Binary String",
		},
	})

	register(&Skill{
		ID:          "simulation",
		Name:        "Simulation",
		Domain:      "algorithm-patterns",
		Description: "Step-by-step execution following problem rules",
		Facets: []string{
			"state representation (what to track)",
			"transition rules (how state changes each step)",
			"termination conditions",
			"optimization (detect cycles, skip redundant steps)",
			"matrix traversal patterns (spiral, diagonal)",
		},
		ExampleProblems: []string{
			"Spiral matrix",
			"Game of life",
			"Robot bounded in circle",
			"Asteroid collision",
			"Design snake game",
		},
	})

	register(&Skill{
		ID:          "counting",
		Name:        "Counting & Combinatorics",
		Domain:      "algorithm-patterns",
		Description: "Counting arrangements, combinations, and paths",
		Facets: []string{
			"permutations vs combinations formula",
			"Pascal's triangle (nCr computation)",
			"inclusion-exclusion principle",
			"counting paths in grid (DP approach)",
			"Catalan numbers (valid parentheses, BST count)",
		},
		ExampleProblems: []string{
			"Unique paths",
			"Unique paths II (with obstacles)",
			"Unique binary search trees",
			"Letter combinations of phone number",
			"Count sorted vowel strings",
		},
	})

	register(&Skill{
		ID:          "design-ds",
		Name:        "Data Structure Design",
		Domain:      "algorithm-patterns",
		Description: "Implementing data structures with specific constraints",
		Facets: []string{
			"combining multiple DS (hashmap + linked list for LRU)",
			"amortized analysis (when occasional expensive ops are OK)",
			"lazy vs eager computation",
			"iterator design (hasNext, next pattern)",
			"handling edge cases (empty, single element)",
		},
		ExampleProblems: []string{
			"LRU cache",
			"LFU cache",
			"Min stack",
			"Design Twitter",
			"Insert delete getRandom O(1)",
		},
	})

	register(&Skill{
		ID:          "divide-and-conquer",
		Name:        "Divide and Conquer",
		Domain:      "algorithm-patterns",
		Description: "Breaking problems into subproblems, solving recursively, combining results",
		Facets: []string{
			"pattern: divide → conquer → combine",
			"merge sort (split, sort halves, merge)",
			"quick select (partition to find kth element in O(n) avg)",
			"closest pair of points (geometric D&C)",
			"recurrence relations (Master theorem basics)",
		},
		ExampleProblems: []string{
			"Merge sort",
			"Kth largest element (quick select)",
			"Count of range sum",
			"Median of two sorted arrays",
			"Maximum subarray (D&C approach)",
		},
	})

	register(&Skill{
		ID:          "line-sweep",
		Name:        "Line Sweep",
		Domain:      "algorithm-patterns",
		Description: "Processing events in sorted order along an axis",
		Facets: []string{
			"event representation (start/end points with type)",
			"sorting events (by position, then by type for ties)",
			"maintaining active set (what's currently intersecting)",
			"counting overlaps (track entry/exit)",
			"combining with other DS (heap, balanced BST)",
		},
		ExampleProblems: []string{
			"The skyline problem",
			"Meeting rooms II (min rooms needed)",
			"Rectangle area II",
			"My calendar II",
			"Employee free time",
		},
	})

	register(&Skill{
		ID:          "reservoir-sampling",
		Name:        "Reservoir Sampling & Randomization",
		Domain:      "algorithm-patterns",
		Description: "Random selection from streams and shuffling algorithms",
		Facets: []string{
			"reservoir sampling (select k items from unknown-size stream)",
			"why it works (probability proof)",
			"Fisher-Yates shuffle (unbiased permutation)",
			"random selection with weights",
			"sampling without replacement",
		},
		ExampleProblems: []string{
			"Linked list random node",
			"Random pick index",
			"Random pick with weight",
			"Shuffle an array",
			"Random point in non-overlapping rectangles",
		},
	})

	register(&Skill{
		ID:          "game-theory",
		Name:        "Game Theory",
		Domain:      "algorithm-patterns",
		Description: "Optimal play in two-player games",
		Facets: []string{
			"minimax (maximize own score, minimize opponent's)",
			"winning vs losing positions (work backwards)",
			"nim game (XOR of pile sizes)",
			"Sprague-Grundy theorem (game states as numbers)",
			"alpha-beta pruning (optimization for minimax)",
		},
		ExampleProblems: []string{
			"Nim game",
			"Stone game",
			"Predict the winner",
			"Can I win",
			"Cat and mouse",
		},
	})

	// System Design
	register(&Skill{
		ID:          "load-balancing",
		Name:        "Load Balancing",
		Domain:      "system-design",
		Description: "Distributing traffic across servers",
		Facets: []string{
			"algorithms (round robin, weighted, least connections)",
			"health checks and failover",
			"sticky sessions (when needed, trade-offs)",
			"L4 vs L7 load balancing (TCP vs HTTP awareness)",
			"DNS load balancing (geographic, latency-based)",
		},
		ExampleProblems: []string{
			"Design a load balancer",
			"Handle server failures gracefully",
			"Session affinity requirements",
			"Geographic load distribution",
		},
	})

	register(&Skill{
		ID:          "consistent-hashing",
		Name:        "Consistent Hashing",
		Domain:      "system-design",
		Description: "Hash-based distribution with minimal redistribution on changes",
		Facets: []string{
			"problem with modulo hashing (N changes → most keys move)",
			"hash ring concept (map keys and nodes to positions on ring)",
			"key assignment (walk clockwise to find responsible node)",
			"virtual nodes (multiple positions per physical node, better distribution)",
			"node add/remove (only keys between neighbors move, ~1/N keys)",
			"use cases: distributed caches, database sharding, CDNs, load balancing",
			"implementations: Cassandra, DynamoDB, memcached, Chord DHT",
		},
		ExampleProblems: []string{
			"Why does adding a node only move ~1/N keys?",
			"How do virtual nodes help with load balancing?",
			"Design a distributed cache with consistent hashing",
			"How would you handle hotspots with consistent hashing?",
		},
	})

	register(&Skill{
		ID:          "caching",
		Name:        "Caching",
		Domain:      "system-design",
		Description: "Storing data for faster access",
		Facets: []string{
			"cache eviction policies (LRU, LFU, FIFO)",
			"cache invalidation strategies",
			"read-through vs write-through vs write-behind",
			"cache aside pattern",
			"distributed caching (Redis, Memcached)",
		},
		ExampleProblems: []string{
			"Design a cache system",
			"Cache invalidation for social feed",
			"CDN caching strategy",
			"Multi-level caching",
		},
	})

	register(&Skill{
		ID:          "sql-vs-nosql",
		Name:        "SQL vs NoSQL",
		Domain:      "system-design",
		Description: "Choosing between relational and non-relational databases",
		Facets: []string{
			"relational (SQL): ACID, joins, schema enforcement, vertical scaling",
			"document stores (MongoDB): flexible schema, nested data, horizontal scaling",
			"key-value stores (Redis, DynamoDB): simple lookups, high throughput, caching",
			"wide-column stores (Cassandra, HBase): time-series, write-heavy, column families",
			"graph databases (Neo4j): relationships, traversals, social networks",
			"when to use SQL (complex queries, transactions, strong consistency)",
			"when to use NoSQL (scale, flexibility, specific access patterns)",
		},
		ExampleProblems: []string{
			"Would you use SQL or NoSQL for an e-commerce product catalog?",
			"What database would you choose for a social network's friend graph?",
			"How would you store time-series metrics at scale?",
			"When would you combine SQL and NoSQL in the same system?",
		},
	})

	register(&Skill{
		ID:          "database-indexing",
		Name:        "Database Indexing",
		Domain:      "system-design",
		Description: "Index structures and query optimization",
		Facets: []string{
			"what an index is (data structure for fast lookups, trade-off: read vs write)",
			"B-tree indexes (balanced tree, O(log n) lookups, range queries)",
			"hash indexes (O(1) exact match, no range queries)",
			"composite indexes (multi-column, leftmost prefix rule)",
			"covering indexes (index contains all needed columns, no table lookup)",
			"index selectivity (high cardinality = more selective = better)",
			"when NOT to index (small tables, low selectivity, write-heavy)",
			"query planning (EXPLAIN, index selection, full table scan)",
		},
		ExampleProblems: []string{
			"How would you optimize a slow query?",
			"When would a composite index help vs hurt?",
			"Why might adding an index make writes slower?",
			"How do you decide which columns to index?",
		},
	})

	register(&Skill{
		ID:          "acid-transactions",
		Name:        "ACID & Transactions",
		Domain:      "system-design",
		Description: "Database transaction guarantees and isolation levels",
		Facets: []string{
			"Atomicity (all or nothing, rollback on failure)",
			"Consistency (valid state to valid state, constraints enforced)",
			"Isolation (concurrent transactions don't interfere)",
			"Durability (committed data survives crashes, WAL)",
			"isolation levels (read uncommitted, read committed, repeatable read, serializable)",
			"phenomena (dirty reads, non-repeatable reads, phantom reads)",
			"distributed transactions (2PC, Saga pattern, eventual consistency)",
			"trade-offs (stronger isolation = lower concurrency)",
		},
		ExampleProblems: []string{
			"What isolation level would you use for a banking system?",
			"How would you handle transactions across microservices?",
			"What's the difference between 2PC and Saga?",
			"When is eventual consistency acceptable?",
		},
	})

	register(&Skill{
		ID:          "database-sharding",
		Name:        "Database Sharding",
		Domain:      "system-design",
		Description: "Horizontal partitioning of data",
		Facets: []string{
			"sharding strategies (hash, range, directory)",
			"shard key selection",
			"cross-shard queries",
			"rebalancing shards",
			"trade-offs (complexity vs scalability)",
		},
		ExampleProblems: []string{
			"Design a sharded database",
			"Handle hotspots",
			"Shard a social network's data",
			"Cross-shard transactions",
		},
	})

	register(&Skill{
		ID:          "message-queues",
		Name:        "Message Queues",
		Domain:      "system-design",
		Description: "Async communication between services",
		Facets: []string{
			"pub/sub vs point-to-point",
			"delivery guarantees (at-least-once, at-most-once, exactly-once)",
			"ordering guarantees",
			"dead letter queues",
			"backpressure handling",
		},
		ExampleProblems: []string{
			"Design a notification system",
			"Order processing pipeline",
			"Event-driven architecture",
			"Handle message failures",
		},
	})

	register(&Skill{
		ID:          "cap-theorem",
		Name:        "CAP Theorem",
		Domain:      "system-design",
		Description: "Consistency, Availability, Partition tolerance trade-offs",
		Facets: []string{
			"what each property means",
			"why you can only have 2 of 3",
			"CP vs AP systems (examples)",
			"eventual consistency",
			"real-world trade-off decisions",
		},
		ExampleProblems: []string{
			"Design a distributed key-value store",
			"Choose consistency model for a banking app",
			"Handle network partitions",
			"Eventual consistency in social feeds",
		},
	})

	register(&Skill{
		ID:          "rate-limiting",
		Name:        "Rate Limiting",
		Domain:      "system-design",
		Description: "Controlling request rates to protect systems",
		Facets: []string{
			"algorithms (token bucket, leaky bucket, fixed window, sliding window)",
			"distributed rate limiting",
			"rate limit by user vs IP vs API key",
			"handling rate limit exceeded",
			"graceful degradation",
		},
		ExampleProblems: []string{
			"Design a rate limiter",
			"API throttling strategy",
			"Prevent abuse while allowing bursts",
			"Distributed rate limiting across servers",
		},
	})

	register(&Skill{
		ID:          "database-replication",
		Name:        "Database Replication",
		Domain:      "system-design",
		Description: "Copying data across multiple database nodes",
		Facets: []string{
			"single-leader vs multi-leader vs leaderless",
			"synchronous vs asynchronous replication",
			"replication lag and read-after-write consistency",
			"failover and leader election",
			"conflict resolution in multi-leader setups",
		},
		ExampleProblems: []string{
			"Design a replicated database",
			"Handle leader failure and failover",
			"Read-your-writes consistency guarantee",
			"Multi-region database deployment",
		},
	})

	register(&Skill{
		ID:          "api-gateway",
		Name:        "API Gateway",
		Domain:      "system-design",
		Description: "Single entry point for client requests to backend services",
		Facets: []string{
			"request routing to microservices",
			"authentication and authorization",
			"rate limiting and throttling at edge",
			"protocol translation (REST to gRPC)",
			"API versioning and backward compatibility",
		},
		ExampleProblems: []string{
			"Design an API gateway for microservices",
			"Handle authentication at the edge",
			"API versioning strategy",
			"Circuit breaker pattern integration",
		},
	})

	register(&Skill{
		ID:          "cdn",
		Name:        "Content Delivery Networks",
		Domain:      "system-design",
		Description: "Distributed network for serving content closer to users",
		Facets: []string{
			"edge caching and cache invalidation",
			"origin servers and pull vs push CDN",
			"geographic distribution and latency reduction",
			"cache hit ratio optimization",
			"dynamic vs static content caching",
		},
		ExampleProblems: []string{
			"Design a CDN",
			"Video streaming architecture",
			"Cache invalidation strategy for dynamic content",
			"Multi-region content delivery",
		},
	})

	register(&Skill{
		ID:          "distributed-coordination",
		Name:        "Distributed Coordination",
		Domain:      "system-design",
		Description: "Consensus and coordination in distributed systems",
		Facets: []string{
			"leader election algorithms",
			"distributed locks and fencing tokens",
			"consensus protocols (Paxos, Raft basics)",
			"service discovery and registration",
			"configuration management (ZooKeeper, etcd)",
		},
		ExampleProblems: []string{
			"Design a distributed lock service",
			"Leader election for database cluster",
			"Service discovery for microservices",
			"Distributed configuration management",
		},
	})

	register(&Skill{
		ID:          "search-systems",
		Name:        "Search Systems",
		Domain:      "system-design",
		Description: "Full-text search, indexing, and query systems",
		Facets: []string{
			"inverted index (term → document mapping)",
			"tokenization and text processing (stemming, stopwords)",
			"ranking algorithms (TF-IDF, BM25 basics)",
			"autocomplete and typeahead (trie + ranking)",
			"scaling search (sharding by term vs document)",
		},
		ExampleProblems: []string{
			"Design a search engine",
			"Design autocomplete system",
			"Design a document search service",
			"Design a product search for e-commerce",
		},
	})

	register(&Skill{
		ID:          "real-time-systems",
		Name:        "Real-Time Systems",
		Domain:      "system-design",
		Description: "Push-based communication and live updates",
		Facets: []string{
			"WebSockets (persistent bidirectional connection)",
			"long polling (simulated push over HTTP)",
			"Server-Sent Events (server push, simpler than WS)",
			"presence systems (online/offline status)",
			"fan-out strategies (push vs pull vs hybrid)",
		},
		ExampleProblems: []string{
			"Design a chat application",
			"Design a live sports scoreboard",
			"Design a collaborative document editor",
			"Design a notification system with live updates",
		},
	})

	register(&Skill{
		ID:          "storage-systems",
		Name:        "Storage Systems",
		Domain:      "system-design",
		Description: "Object storage, file systems, and data persistence",
		Facets: []string{
			"object storage (S3-style: buckets, keys, metadata)",
			"block vs file vs object storage trade-offs",
			"data durability (replication, erasure coding)",
			"tiered storage (hot/warm/cold)",
			"consistency models in distributed storage",
		},
		ExampleProblems: []string{
			"Design a file storage service (Dropbox)",
			"Design an image hosting service",
			"Design a video storage and streaming service",
			"Design a backup system",
		},
	})

	register(&Skill{
		ID:          "observability",
		Name:        "Observability",
		Domain:      "system-design",
		Description: "Monitoring, logging, tracing, and alerting",
		Facets: []string{
			"three pillars: logs, metrics, traces",
			"log aggregation (ELK stack, structured logging)",
			"metrics collection (counters, gauges, histograms)",
			"distributed tracing (trace IDs, spans)",
			"alerting strategies (thresholds, anomaly detection)",
		},
		ExampleProblems: []string{
			"Design a logging infrastructure",
			"Design a metrics collection system",
			"Design a distributed tracing system",
			"Design an alerting system",
		},
	})

	register(&Skill{
		ID:          "tcp-udp-networking",
		Name:        "TCP, UDP, and Network Fundamentals",
		Domain:      "system-design",
		Description: "Transport layer protocols and networking basics for system design",
		Facets: []string{
			"TCP fundamentals (connection-oriented, reliable, ordered, flow control)",
			"TCP handshake (3-way: SYN, SYN-ACK, ACK; teardown: FIN)",
			"TCP congestion control (slow start, AIMD, congestion window)",
			"UDP fundamentals (connectionless, unreliable, no ordering, low overhead)",
			"TCP vs UDP trade-offs (reliability vs latency, gaming/video vs web)",
			"HTTP over TCP (why HTTP/1.1 and HTTP/2 use TCP, head-of-line blocking)",
			"QUIC/HTTP3 (UDP-based, multiplexed streams, 0-RTT)",
			"when to use each (video streaming→UDP, API calls→TCP, real-time games→UDP)",
		},
		ExampleProblems: []string{
			"Why does video conferencing use UDP instead of TCP?",
			"How does TCP ensure reliable delivery?",
			"What causes head-of-line blocking in HTTP/2?",
			"When would you choose QUIC over TCP?",
		},
	})

	register(&Skill{
		ID:          "realtime-communication",
		Name:        "Real-Time Communication Patterns",
		Domain:      "system-design",
		Description: "Comparing polling, long-polling, SSE, and WebSockets for real-time updates",
		Facets: []string{
			"simple polling (client pulls at intervals, high latency, wasteful)",
			"long polling (hold request open until data, better latency, connection overhead)",
			"Server-Sent Events (SSE) (server push over HTTP, unidirectional, auto-reconnect)",
			"WebSockets (HTTP upgrade → persistent TCP, bidirectional, low latency)",
			"trade-offs: latency, scalability, firewall compatibility, mobile battery",
			"when to use each (notifications→SSE, chat→WS, dashboards→polling/SSE)",
		},
		ExampleProblems: []string{
			"How would you implement live sports scores?",
			"How would you push notifications to a web app?",
			"How would you build a collaborative document editor?",
			"How would you implement a stock ticker?",
		},
	})

	register(&Skill{
		ID:          "change-data-capture",
		Name:        "Change Data Capture (CDC)",
		Domain:      "system-design",
		Description: "Capturing and propagating database changes to downstream systems",
		Facets: []string{
			"what CDC is (streaming database changes as events)",
			"log-based CDC (reading database WAL/binlog, e.g., Debezium)",
			"trigger-based CDC (database triggers write to change table)",
			"query-based CDC (polling for changes via timestamp/version)",
			"use cases: cache invalidation, search index sync, event sourcing, data replication",
			"trade-offs: latency, consistency, schema evolution, operational complexity",
		},
		ExampleProblems: []string{
			"How would you keep Elasticsearch in sync with PostgreSQL?",
			"How would you invalidate cache when database changes?",
			"How would you replicate data across microservices?",
			"How would you build an audit log for all database changes?",
		},
	})

	register(&Skill{
		ID:          "presigned-urls",
		Name:        "Pre-signed URLs",
		Domain:      "system-design",
		Description: "Secure, temporary direct access to cloud storage objects",
		Facets: []string{
			"what pre-signed URLs are (time-limited signed URLs for direct S3/GCS access)",
			"upload flow (client requests URL from server, uploads directly to storage)",
			"download flow (server generates URL, client downloads directly)",
			"security: expiration time, IP restrictions, content-type limits",
			"benefits: offload bandwidth from app servers, reduce latency",
			"use cases: large file uploads, CDN origin, mobile apps, browser uploads",
		},
		ExampleProblems: []string{
			"How would you handle large file uploads without overloading your servers?",
			"How would you let users download files securely without proxying through your API?",
			"How would you implement resumable uploads for mobile apps?",
			"How would you design a file sharing system like Dropbox?",
		},
	})

	register(&Skill{
		ID:          "ci-cd-systems",
		Name:        "CI/CD Systems",
		Domain:      "system-design-practical",
		Description: "Continuous integration and deployment workflow systems",
		Facets: []string{
			"event triggering (webhooks, polling, push events from VCS)",
			"workflow orchestration (YAML parsing, job DAG, dependency resolution)",
			"distributed execution (worker pools, containerized runners, resource isolation)",
			"artifact management (build outputs, caching, storage)",
			"real-time observability (log streaming, status updates, webhooks)",
			"scalability (handling 10M+ repos, burst traffic, job queuing)",
		},
		ExampleProblems: []string{
			"Design GitHub Actions from scratch",
			"Design a CI pipeline for monorepo at scale",
			"Design a self-hosted runner infrastructure",
			"Design a build artifact caching system",
		},
	})

	register(&Skill{
		ID:          "online-chess-platform",
		Name:        "Online Chess Platform",
		Domain:      "system-design-practical",
		Description: "Low-latency multiplayer game system with real-time state sync",
		Facets: []string{
			"matchmaking (quick pairings, rating buckets, timeout handling)",
			"real-time game loop (move submission, validation, turn enforcement)",
			"state synchronization (WebSockets, ordered events, reconnect recovery)",
			"consistency and fairness (clock handling, idempotent moves, anti-cheat checks)",
			"persistence and analytics (game history, PGN/event logs, rating updates)",
		},
		ExampleProblems: []string{
			"Design Chess.com/Lichess style live play",
			"Handle reconnect during an in-progress game",
			"Design rating updates after match completion",
			"Scale real-time spectators for popular games",
		},
	})

	register(&Skill{
		ID:          "messenger-chat-system",
		Name:        "Messenger / Chat System",
		Domain:      "system-design-practical",
		Description: "Real-time 1:1 messaging with presence, receipts, and multi-device sync",
		Facets: []string{
			"delivery semantics (at-least-once, deduplication, retries)",
			"ordering (server sequence per conversation, shard ownership)",
			"real-time presence and low-latency fan-out",
			"multi-device synchronization (per-device cursors and reconciliation)",
			"storage lifecycle (high write throughput, retention, compliance deletes)",
		},
		ExampleProblems: []string{
			"Design WhatsApp/Facebook Messenger",
			"Guarantee message ordering across flaky networks",
			"Design presence and read receipts at scale",
			"Build multi-device history sync with offline retry",
		},
	})

	register(&Skill{
		ID:          "design-twitter",
		Name:        "Design Twitter",
		Domain:      "system-design-practical",
		Description: "Social media platform with feeds, posts, follows, and timeline ranking",
		Facets: []string{
			"feed generation (fanout-on-write vs fanout-on-read)",
			"timeline ranking and personalization",
			"handling celebrity accounts (millions of followers)",
			"caching strategy (user timeline, home feed)",
			"real-time updates and notifications",
		},
		ExampleProblems: []string{
			"Twitter/X",
			"Facebook News Feed",
			"Instagram Feed",
			"LinkedIn Feed",
		},
	})

	register(&Skill{
		ID:          "design-uber",
		Name:        "Design Uber",
		Domain:      "system-design-practical",
		Description: "Ride-sharing platform with real-time matching, routing, and payments",
		Facets: []string{
			"real-time location tracking and geospatial indexing",
			"driver-rider matching algorithm",
			"surge pricing and demand prediction",
			"ETA calculation and routing",
			"payment processing and fraud detection",
		},
		ExampleProblems: []string{
			"Uber/Lyft",
			"DoorDash/Instacart",
			"Yelp (nearby search)",
			"Google Maps (routing)",
		},
	})

	register(&Skill{
		ID:          "design-dropbox",
		Name:        "Design Dropbox",
		Domain:      "system-design-practical",
		Description: "File storage and sync service with versioning and collaboration",
		Facets: []string{
			"chunking and deduplication",
			"sync protocol (delta sync, conflict resolution)",
			"metadata vs content storage separation",
			"file versioning and history",
			"sharing and permissions",
		},
		ExampleProblems: []string{
			"Dropbox/Google Drive",
			"OneDrive",
			"iCloud Drive",
			"Box",
		},
	})

	// LeetCode Patterns (Archetypes)
	register(&Skill{
		ID:          "cooldown-scheduling",
		Name:        "Cooldown Scheduling",
		Domain:      "leetcode-patterns",
		Description: "Problems where items must be placed with minimum spacing constraints",
		Facets: []string{
			"recognition (spacing/cooldown constraint in problem)",
			"why heap (need max frequency available item)",
			"why queue (FIFO cooldown tracking)",
			"greedy correctness (high frequency first avoids deadlock)",
			"complexity analysis (O(n log k) where k is unique items)",
		},
		ExampleProblems: []string{
			"Task Scheduler",
			"Rearrange String K Distance Apart",
			"Reorganize String",
		},
	})

	register(&Skill{
		ID:          "two-heaps-median",
		Name:        "Two Heaps for Median",
		Domain:      "leetcode-patterns",
		Description: "Maintain running median using two heaps",
		Facets: []string{
			"recognition (need median of dynamic data)",
			"structure (max-heap for lower half, min-heap for upper half)",
			"balancing invariant (sizes differ by at most 1)",
			"median retrieval (O(1) from heap tops)",
			"insertion logic (which heap, then rebalance)",
		},
		ExampleProblems: []string{
			"Find Median from Data Stream",
			"Sliding Window Median",
			"IPO (maximize capital)",
		},
	})

	register(&Skill{
		ID:          "monotonic-stack-optimization",
		Name:        "Monotonic Stack Optimization",
		Domain:      "leetcode-patterns",
		Description: "Use stack to find next greater/smaller in O(n)",
		Facets: []string{
			"recognition (next greater/smaller element pattern)",
			"stack invariant (monotonically increasing or decreasing)",
			"what triggers pop (element breaks monotonic property)",
			"what to compute on pop (width, area, span)",
			"handling leftovers (elements remaining in stack)",
		},
		ExampleProblems: []string{
			"Largest Rectangle in Histogram",
			"Trapping Rain Water",
			"Daily Temperatures",
			"Next Greater Element",
		},
	})

	register(&Skill{
		ID:          "sliding-window-hash",
		Name:        "Sliding Window + Hash Map",
		Domain:      "leetcode-patterns",
		Description: "Track window contents with hash map for O(1) lookups",
		Facets: []string{
			"recognition (substring/subarray with constraint)",
			"window state (hash map tracking counts or positions)",
			"expand condition (when to grow window)",
			"shrink condition (when window violates constraint)",
			"answer extraction (min/max window seen)",
		},
		ExampleProblems: []string{
			"Minimum Window Substring",
			"Longest Substring Without Repeating Characters",
			"Longest Repeating Character Replacement",
			"Permutation in String",
		},
	})

	register(&Skill{
		ID:          "bfs-complex-state",
		Name:        "BFS with Complex State",
		Domain:      "leetcode-patterns",
		Description: "BFS where state includes more than just position",
		Facets: []string{
			"recognition (shortest path with additional constraints)",
			"state definition (position + extra info like keys, steps, fuel)",
			"visited tracking (must track full state, not just position)",
			"state encoding (tuple, string, or bit manipulation)",
			"pruning opportunities (avoid redundant states)",
		},
		ExampleProblems: []string{
			"Open the Lock",
			"Word Ladder",
			"Shortest Path with Obstacles Elimination",
			"Shortest Path to Get All Keys",
		},
	})

	register(&Skill{
		ID:          "dp-on-intervals",
		Name:        "DP on Intervals",
		Domain:      "leetcode-patterns",
		Description: "DP where state is an interval [i,j]",
		Facets: []string{
			"recognition (optimal substructure on contiguous ranges)",
			"state definition (dp[i][j] = optimal for interval [i,j])",
			"transition (try all split points k in [i,j])",
			"iteration order (by interval length, small to large)",
			"base cases (single element or empty intervals)",
		},
		ExampleProblems: []string{
			"Burst Balloons",
			"Matrix Chain Multiplication",
			"Minimum Cost to Merge Stones",
			"Strange Printer",
		},
	})

	register(&Skill{
		ID:          "binary-search-on-answer",
		Name:        "Binary Search on Answer",
		Domain:      "leetcode-patterns",
		Description: "Binary search over possible answer values",
		Facets: []string{
			"recognition (minimize maximum or maximize minimum)",
			"monotonicity (if answer X works, X+1 also works, or vice versa)",
			"predicate function (can we achieve this answer?)",
			"search space bounds (min and max possible answers)",
			"answer extraction (first/last valid value)",
		},
		ExampleProblems: []string{
			"Koko Eating Bananas",
			"Split Array Largest Sum",
			"Capacity To Ship Packages",
			"Minimize Max Distance to Gas Station",
		},
	})

	register(&Skill{
		ID:          "topological-ordering",
		Name:        "Topological Order Applications",
		Domain:      "leetcode-patterns",
		Description: "Use topological sort for dependency resolution",
		Facets: []string{
			"recognition (dependencies, prerequisites, ordering constraints)",
			"cycle detection (impossible if cycle exists)",
			"Kahn's algorithm (BFS with indegree tracking)",
			"multiple valid orderings (lexicographically smallest)",
			"counting orderings (DP on topological order)",
		},
		ExampleProblems: []string{
			"Course Schedule I & II",
			"Alien Dictionary",
			"Sequence Reconstruction",
			"Parallel Courses",
		},
	})

	register(&Skill{
		ID:          "union-find-patterns",
		Name:        "Union-Find Patterns",
		Domain:      "leetcode-patterns",
		Description: "Dynamic connectivity and component tracking",
		Facets: []string{
			"recognition (grouping, connectivity, equivalence)",
			"path compression (find optimization)",
			"union by rank/size (union optimization)",
			"component tracking (count, size, or properties)",
			"online vs offline (process queries in order vs sort first)",
		},
		ExampleProblems: []string{
			"Accounts Merge",
			"Redundant Connection",
			"Number of Islands II",
			"Smallest String With Swaps",
		},
	})

	register(&Skill{
		ID:          "tree-path-problems",
		Name:        "Tree Path Problems",
		Domain:      "leetcode-patterns",
		Description: "Problems involving paths in trees",
		Facets: []string{
			"recognition (path sum, diameter, LCA)",
			"path types (root-to-leaf vs any-to-any)",
			"DFS state (what to pass down vs return up)",
			"combining child results (max path through node)",
			"global vs local answer (update global during DFS)",
		},
		ExampleProblems: []string{
			"Binary Tree Maximum Path Sum",
			"Path Sum III",
			"Diameter of Binary Tree",
			"Longest Univalue Path",
		},
	})

	register(&Skill{
		ID:          "prefix-sum-tricks",
		Name:        "Prefix Sum Tricks",
		Domain:      "leetcode-patterns",
		Description: "Prefix sums with hash map for subarray queries",
		Facets: []string{
			"recognition (subarray sum equals target)",
			"prefix sum + hash map (count subarrays with sum K)",
			"modular arithmetic (divisibility conditions)",
			"prefix XOR (subarray XOR problems)",
			"2D prefix sums (matrix region queries)",
		},
		ExampleProblems: []string{
			"Subarray Sum Equals K",
			"Contiguous Array",
			"Subarray Sums Divisible by K",
			"Find Pivot Index",
		},
	})

	register(&Skill{
		ID:          "greedy-intervals",
		Name:        "Greedy Interval Scheduling",
		Domain:      "leetcode-patterns",
		Description: "Greedy algorithms on intervals",
		Facets: []string{
			"recognition (intervals with selection/removal)",
			"sorting strategy (by start, end, or both)",
			"greedy choice (earliest end time, or latest start)",
			"proof technique (exchange argument)",
			"heap for tracking (overlapping intervals count)",
		},
		ExampleProblems: []string{
			"Non-overlapping Intervals",
			"Meeting Rooms II",
			"Minimum Number of Arrows",
			"Insert Interval",
		},
	})
}
