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
	return []string{"data-structures", "algorithm-patterns", "system-design"}
}

// Domain short names
var DomainMap = map[string]string{
	"ds":                 "data-structures",
	"data-structures":    "data-structures",
	"algo":               "algorithm-patterns",
	"algorithms":         "algorithm-patterns",
	"algorithm-patterns": "algorithm-patterns",
	"sys":                "system-design",
	"system":             "system-design",
	"system-design":      "system-design",
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
		Name:        "Segment Trees",
		Domain:      "data-structures",
		Description: "Tree structure for efficient range queries and point updates",
		Facets: []string{
			"structure (complete binary tree over array)",
			"build O(n), query/update O(log n)",
			"lazy propagation for range updates",
			"application (range sum, min/max, GCD)",
			"when to use vs prefix sum or Fenwick tree",
		},
		ExampleProblems: []string{
			"Range sum query - mutable",
			"Count of range sum",
			"Range minimum query",
			"Falling squares",
			"My calendar III",
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
		ID:          "shortest-path",
		Name:        "Shortest Path Algorithms",
		Domain:      "algorithm-patterns",
		Description: "Finding optimal paths in weighted graphs",
		Facets: []string{
			"Dijkstra's algorithm (non-negative weights, greedy)",
			"Bellman-Ford (handles negative weights)",
			"when to use BFS vs Dijkstra",
			"negative cycle detection",
			"priority queue optimization for Dijkstra",
		},
		ExampleProblems: []string{
			"Network delay time",
			"Cheapest flights within K stops",
			"Path with minimum effort",
			"Swim in rising water",
			"Find the city with smallest number of neighbors",
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
			"consistent hashing",
			"health checks",
			"sticky sessions (when needed, trade-offs)",
			"L4 vs L7 load balancing",
		},
		ExampleProblems: []string{
			"Design a load balancer",
			"Handle server failures gracefully",
			"Session affinity requirements",
			"Geographic load distribution",
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
}
