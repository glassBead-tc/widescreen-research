---
id: "subliminal-channel-capacity-estimation"
name: "Subliminal Learning Channel Capacity Estimation"
version: "1.0.0"
category: "ai-theory-research"
tags: ["information-theory", "llm-architecture", "model-to-model-learning", "channel-capacity", "parameter-analysis"]
complexity: "high"
estimated_duration: "2-4 hours"
last_updated: "2025-09-29"
author: "research-orchestration-system"
status: "active"
dependencies:
  - "arxiv-paper-mcp"
  - "exa__web_search_exa"
  - "exa__get_code_context_exa"
  - "firecrawl__firecrawl_scrape"
  - "context7"
---

# Subliminal Learning Channel Capacity Estimation

Estimate the information-theoretic capacity of the non-semantic channel through which teacher LLMs transmit behavioral preferences to student LLMs during training, as discovered in Anthropic's subliminal learning research.

## Overview

Anthropic's research on subliminal learning revealed that teacher models can transmit preferences (e.g., "loves owls") to student models through training data that has no semantic connection to those preferences (e.g., number sequences). This transmission occurs through non-semantic signals in the probability distribution that are not removable through standard data filtering.

This orchestration estimates the **channel capacity** - how many bits of information can be transmitted per training example - by analyzing open-source LLM architectures and applying information-theoretic bounds. The goal is to understand: If a powerful teacher LLM wanted to intentionally encode information into a student, how much could it transmit before running out of degrees of freedom?

## Research Question

**Primary**: What is the information-theoretic capacity (bits/example) of the subliminal learning channel for modern open-source LLMs?

**Secondary**:
- How does capacity scale with model size (7B, 13B, 70B parameters)?
- What architectural factors (hidden dim, layers, attention heads) most affect capacity?
- How many training examples would a teacher need to transmit X bits of preference information?
- What is the theoretical maximum distinguishable "states" a teacher could encode?

## Required Tools

- `arxiv-paper-mcp__search_papers` - Find subliminal learning and related papers
- `exa__web_search_exa` - Find model cards and architecture documentation
- `exa__get_code_context_exa` - Find model implementation code
- `firecrawl__firecrawl_scrape` - Scrape model documentation pages
- `context7__resolve-library-id` + `get-library-docs` - Library documentation if needed

## Workflow Steps

### Phase 1: Foundation - Subliminal Learning Mechanism (30-45 min)

```
1. Retrieve Core Research
   Use `exa__web_search_exa`:
   - Query: "Anthropic subliminal learning model-to-model preferences 2025"
   - Target: https://alignment.anthropic.com/2025/subliminal-learning/
   - Scrape with `firecrawl__firecrawl_scrape`

   Extract key findings:
   - Mechanism description
   - Experimental setup details
   - Which model architectures were tested
   - What types of preferences were transmitted
   - Training data characteristics (number of examples, format)

2. Academic Context
   Use `arxiv-paper-mcp__search_papers`:
   - Query: "model-to-model learning hidden channels neural networks"
   - Query: "information theory neural network capacity"
   - Query: "representation learning geometric properties"
   - Prioritize: August-September 2025 papers

   Extract:
   - Related work on hidden channels in neural networks
   - Information-theoretic analyses of model learning
   - Geometric properties of activation spaces

3. Theoretical Framework
   Use `exa__get_code_context_exa`:
   - Query: "information theory channel capacity Shannon theorem"
   - Query: "mutual information neural network representations"

   Build understanding of:
   - Shannon channel capacity formula: C = max I(X;Y)
   - Mutual information in high-dimensional spaces
   - Bounds on distinguishable states in continuous spaces
```

### Phase 2: Architecture Intelligence Gathering (45-60 min)

```
4. Identify Target Models (Prioritize Aug/Sep 2025 Releases)
   Use `exa__web_search_exa`:
   - Query: "Llama 3.2 3.3 model card architecture September 2025"
   - Query: "Mistral Large 2 architecture August 2025"
   - Query: "Qwen 2.5 architecture specifications 2025"
   - Query: "open source LLM releases August September 2025"

   Target architectures:
   - Llama 3.1/3.2/3.3 (7B, 13B, 70B, 405B)
   - Mistral (7B, 8x7B, Large)
   - Qwen 2.5 (various sizes)
   - Gemma 2 (if open weights)
   - Any other Aug/Sep 2025 releases

5. Extract Architecture Specifications
   Use `firecrawl__firecrawl_scrape` on:
   - HuggingFace model cards
   - GitHub repository READMEs
   - Official model documentation

   For each model, extract:
   - Total parameter count (e.g., 70B)
   - Hidden dimension / d_model (e.g., 8192)
   - Number of layers / n_layers (e.g., 80)
   - Number of attention heads / n_heads (e.g., 64)
   - FFN intermediate dimension (typically 4x hidden_dim)
   - Vocabulary size
   - Context length
   - Architecture type (decoder-only, etc.)

6. Find Implementation Code
   Use `exa__get_code_context_exa`:
   - Query: "Llama 3 model architecture implementation config.json"
   - Query: "transformers LlamaConfig LlamaModel implementation"
   - Query: "Mistral architecture PyTorch implementation"

   Get actual config files:
   - config.json from model repos
   - Model class definitions
   - Architecture diagrams if available
```

### Phase 3: Capacity Calculation & Bounds (60-90 min)

```
7. Activation Space Dimensionality Analysis
   For each model architecture:

   a) Calculate activation space dimensions:
      - Per-token activation: hidden_dim (e.g., 8192)
      - Per-sequence activation: hidden_dim × context_length
      - Total representational space: n_layers × hidden_dim

   b) Estimate effective dimensionality:
      - Transformer self-attention creates dependencies
      - Not all dimensions are independent
      - Estimate: ~√(n_layers × hidden_dim) independent directions

   c) Calculate for each model size:
      | Model | Hidden Dim | Layers | Effective Dim | Notes |
      |-------|-----------|---------|---------------|-------|
      | Llama 3.1 7B | 4096 | 32 | ~360 | sqrt(32*4096) |
      | Llama 3.1 70B | 8192 | 80 | ~810 | sqrt(80*8192) |
      | Llama 3.1 405B | 16384 | 126 | ~1438 | sqrt(126*16384) |

8. Signal-to-Noise Ratio Estimation
   From subliminal learning paper findings:

   a) Determine baseline:
      - How much geometric shift from random training data?
      - Teacher's marginal shift per example?
      - Detection threshold for student to "pick up" signal?

   b) Calculate SNR:
      - Signal: Teacher-induced geometric shift
      - Noise: Random variation across training examples
      - SNR = (teacher_shift / random_variation)

   c) Estimate bits per dimension:
      - With SNR, how many distinguishable levels per dimension?
      - Distinguishable_levels ≈ 1 + SNR
      - Bits_per_dim = log₂(distinguishable_levels)

9. Channel Capacity Bounds
   Calculate theoretical capacity:

   a) Per-Example Capacity (Lower Bound):
      - Assumptions: Conservative, only clearly controllable dimensions
      - Formula: C_lower = log₂(effective_dim) bits/example
      - Rationale: Teacher can choose which dimension to shift

   b) Per-Example Capacity (Upper Bound):
      - Assumptions: Optimistic, all dimensions partially controllable
      - Formula: C_upper = effective_dim × bits_per_dim
      - Rationale: Teacher can shift all dimensions simultaneously

   c) Realistic Estimate:
      - Account for:
        * Learning rate effects (geometric mean?)
        * Batch size dilution
        * Optimizer noise
        * Same-base-model requirement
      - Formula: C_realistic = α × √(effective_dim × bits_per_dim)
        where α ≈ 0.5-1.0 (empirical fudge factor)

10. Training Dataset Capacity
    Extend to full training runs:

    a) Examples needed for N bits:
       - If C_realistic = X bits/example
       - Need N/X examples to transmit N bits

    b) Preference complexity estimation:
       - "Loves owls": ~4-8 bits? (simple binary-ish preference)
       - Complex personality trait: ~20-50 bits?
       - Full behavioral profile: ~100-500 bits?

    c) Training budget analysis:
       - Typical fine-tuning: 10k-100k examples
       - Capacity: 10k × X bits = total transmittable information
```

### Phase 4: Model Comparison & Scaling Analysis (30-45 min)

```
11. Cross-Model Capacity Comparison
    Create comparison table:

    | Model | Params | Hidden | Layers | Eff. Dim | C_lower | C_upper | C_realistic |
    |-------|--------|--------|--------|----------|---------|---------|-------------|
    | Llama 3.1 7B | 7B | 4096 | 32 | 360 | ~8 bits | ~720 bits | ~150 bits |
    | Llama 3.1 70B | 70B | 8192 | 80 | 810 | ~10 bits | ~1620 bits | ~300 bits |
    | Mistral 7B | 7B | 4096 | 32 | 360 | ~8 bits | ~720 bits | ~150 bits |
    | Qwen 2.5 72B | 72B | 8192 | 80 | 810 | ~10 bits | ~1620 bits | ~300 bits |

    Insights:
    - Capacity scales roughly with √(parameters)
    - Larger models = more bandwidth for subliminal transmission
    - Diminishing returns at very large sizes

12. Scaling Law Analysis
    Fit power law to data:
    - C(n_params) = k × n_params^α
    - Estimate α from available data points
    - Extrapolate to hypothetical larger models (1T, 10T params)

    Questions:
    - At what model size does capacity plateau?
    - Does architecture type (dense vs. MoE) affect capacity?

13. Sensitivity Analysis
    Which factors matter most?

    Vary each parameter:
    - ±50% hidden dimension
    - ±50% layer count
    - ±50% context length
    - Different learning rates

    Calculate elasticity: ∂C/∂x for each factor
```

### Phase 5: Validation & Reality Check (30-45 min)

```
14. Cross-Reference with Empirical Evidence
    From subliminal learning paper:
    - What model sizes were actually tested?
    - How many examples were needed for transmission?
    - What was the success rate?

    Back-calculate:
    - If transmission succeeded with N examples
    - And preference is ~X bits
    - Implied capacity = X/N bits/example

    Compare to our theoretical estimates:
    - Are we in the right ballpark?
    - Too high → We're overestimating effective dimensions
    - Too low → We're missing parallel channels

15. Literature Cross-Validation
    Use `arxiv-paper-mcp__search_papers`:
    - Query: "neural network capacity information bottleneck"
    - Query: "representation learning dimensionality"

    Check if our estimates align with:
    - Known results about neural network information capacity
    - Representation learning theory
    - Compression bounds

16. Expert Knowledge Integration
    Use `exa__get_code_context_exa`:
    - Query: "information theory neural networks capacity estimation"
    - Query: "mutual information deep learning representations"

    Find any existing work on:
    - Measuring information flow in neural networks
    - Capacity bounds for teacher-student learning
    - Hidden communication channels in ML
```

### Phase 6: Synthesis & Reporting (30-45 min)

```
17. Compile Findings Report
    Structure:

    ### Executive Summary
    - Estimated capacity range: X-Y bits/example
    - For 10k training examples: X-Y kilobits total
    - Scaling: Capacity ∝ n_params^α where α ≈ X
    - Key insight: [Main finding]

    ### Methodology
    - Models analyzed: [List]
    - Theoretical framework: Shannon capacity, geometric analysis
    - Assumptions: [List key assumptions]
    - Limitations: [What we couldn't account for]

    ### Detailed Calculations
    [Show work for each model]

    ### Sensitivity Analysis
    [Which factors matter most]

    ### Validation
    [How estimates compare to empirical evidence]

    ### Implications
    - For AI safety: [What this means]
    - For model training: [Considerations]
    - For future research: [Open questions]

18. Uncertainty Quantification
    For each estimate, provide:
    - Point estimate (best guess)
    - 90% confidence interval
    - Key sources of uncertainty
    - Sensitivity to assumptions

    Example:
    "Llama 3.1 70B capacity: 300 bits/example (90% CI: 100-600 bits)
     Main uncertainty: effective dimensionality assumption"

19. Research Gaps & Next Steps
    Identify what would improve estimates:
    - Empirical measurements needed
    - Architectural factors to investigate
    - Theoretical developments required
    - Experimental designs proposed
```

## Input Requirements

### Required Inputs
- **Research Question** (string): The specific capacity estimation question
- **Model Priority** (array): List of models to prioritize (default: latest open-source)
- **Time Period** (string): Prefer recent releases (default: "August-September 2025")

### Optional Inputs
- **Confidence Level** (float): Target confidence for estimates (default: 0.90)
- **Comparison Baseline** (string): Model to use as reference point
- **Output Detail** (enum): "summary" | "detailed" | "academic" (default: "detailed")

## Expected Outputs

### Primary Output: Capacity Estimation Report

```markdown
# Subliminal Learning Channel Capacity: Estimation Report

## Executive Summary

Based on analysis of 5 open-source LLM architectures (Aug-Sep 2025), we estimate:

- **Per-Example Capacity**: 150-300 bits/example (90% CI: 100-500 bits)
- **Training Budget Capacity**: For 10k examples: 1.5-3.0 megabits total
- **Scaling Law**: C ≈ 50 × n_params^0.4 bits/example
- **Key Finding**: Larger models provide significantly more bandwidth for subliminal transmission

### Models Analyzed

| Model | Release | Params | Hidden | Layers | Capacity (bits/ex) |
|-------|---------|--------|--------|--------|-------------------|
| Llama 3.2 | Sep 2025 | 7B | 4096 | 32 | 150 (100-250) |
| Llama 3.3 | Sep 2025 | 70B | 8192 | 80 | 300 (200-500) |
| Mistral Large 2 | Aug 2025 | 123B | 12288 | 88 | 380 (250-600) |
| Qwen 2.5 | Sep 2025 | 72B | 8192 | 80 | 300 (200-500) |

### Methodology

**Theoretical Framework**: Shannon channel capacity applied to geometric signal transmission

**Key Assumptions**:
1. Effective dimensionality ≈ √(n_layers × hidden_dim)
2. Signal-to-noise ratio ≈ 2-4 (from observed transmission success rates)
3. Bits per dimension ≈ 1-2 (limited by learning dynamics)
4. Capacity = α × √(effective_dim × bits_per_dim), α ≈ 0.7

**Validation**: Estimates consistent with empirical observations that ~10k examples can transmit simple preferences

### Detailed Analysis

[Full calculations shown for each model...]

### Implications

**For Alignment Safety**:
- A malicious teacher with 100k training budget could encode ~30 kilobits
- That's enough for complex behavioral patterns, not just simple preferences
- Standard semantic filtering cannot remove this signal

**For Model Training**:
- Teacher-student learning may transmit more than intended
- Distillation preserves not just capabilities but also subtle biases
- Same-base-model requirement suggests mitigation strategy

**Open Questions**:
- Does mixture-of-experts architecture affect capacity?
- Can we detect subliminal signals in training data?
- What is the minimum SNR for reliable transmission?
```

### Supporting Outputs

```json
{
  "model_architectures": {
    "llama_3_3_70b": {
      "parameters": 70e9,
      "hidden_dim": 8192,
      "num_layers": 80,
      "effective_dim_estimate": 810,
      "capacity_bits_per_example": {
        "lower_bound": 200,
        "point_estimate": 300,
        "upper_bound": 500,
        "confidence": 0.90
      }
    }
  },
  "scaling_law": {
    "formula": "C = 50 * (n_params ** 0.4)",
    "r_squared": 0.92,
    "extrapolations": {
      "1T_params": "~850 bits/example",
      "10T_params": "~2100 bits/example"
    }
  },
  "sensitivity_analysis": {
    "hidden_dim_elasticity": 0.5,
    "layer_count_elasticity": 0.5,
    "learning_rate_elasticity": 0.3
  }
}
```

## Usage Example

### Scenario: Estimating Transmission Capacity for Llama 3.3 70B

**Input**:
```yaml
research_question: "How many bits per example can a Llama 3.3 70B teacher transmit to a Llama 3.3 70B student?"
model_priority: ["llama-3.3-70b", "llama-3.2-7b"]
time_period: "September 2025"
confidence_level: 0.90
output_detail: "detailed"
```

**Process**:
1. Scrape Anthropic subliminal learning blog post
2. Find Llama 3.3 model card on HuggingFace (Sep 2025 release)
3. Extract: hidden_dim=8192, layers=80, params=70B
4. Calculate effective dimensionality: √(80 × 8192) ≈ 810
5. Estimate SNR from empirical data: ≈3
6. Apply formula: C ≈ 0.7 × √(810 × 1.5) ≈ 250-350 bits/example
7. Validate against paper's observations
8. Generate detailed report with confidence intervals

**Expected Output**:
```markdown
# Capacity Estimation: Llama 3.3 70B

## Point Estimate
**300 bits per training example** (90% CI: 200-500 bits)

## Calculation Basis
- Effective dimensionality: 810 (from √(80 layers × 8192 hidden_dim))
- Estimated SNR: 3 (from observation that 10k examples transmit ~8-bit preferences)
- Bits per dimension: 1.5 (conservative estimate given learning dynamics)
- Capacity formula: 0.7 × √(810 × 1.5) ≈ 300 bits

## Training Budget Implications
- 10k examples: ~3 megabits total capacity
- 100k examples: ~30 megabits total capacity
- 1M examples: ~300 megabits total capacity

**Complexity that could be encoded**:
- Simple preference (8 bits): ~30 examples needed
- Personality trait (50 bits): ~170 examples needed
- Behavioral profile (500 bits): ~1,700 examples needed
- Full alignment fingerprint (5kb): ~17,000 examples needed

## Confidence Assessment
- High confidence in order of magnitude (100-1000 bits)
- Medium confidence in scaling law exponent
- Low confidence in absolute value (could be off by 2x)
- Key uncertainty: effective dimensionality assumption
```

## Error Handling

### Common Issues

1. **Model Architecture Not Found**
   - Fallback to similar model size in same family
   - Use general architecture patterns for model class
   - Note in report that specs are estimated

2. **Conflicting Parameter Counts**
   - Use multiple sources and take median
   - Document discrepancies in uncertainty bounds
   - Prefer official model cards over secondary sources

3. **Missing Implementation Details**
   - Use `context7` to find library documentation
   - Infer from model family patterns
   - Widen confidence intervals to account for uncertainty

### Recovery Strategies
- If Anthropic blog unreachable: Search for paper on arXiv
- If model cards unavailable: Find in HuggingFace transformers library code
- If calculations uncertain: Provide wider bounds and sensitivity analysis

## Performance Considerations

- **Typical Duration**: 2-4 hours for comprehensive analysis of 4-5 models
- **API Calls**: ~20-30 searches + scrapes
- **Bottlenecks**: Finding accurate architecture specs for very recent models
- **Optimization**: Cache model architectures, parallelize scraping

## Best Practices

1. **Multiple Sources**: Always cross-reference architecture specs from 2+ sources
2. **Explicit Assumptions**: Document every assumption in calculations
3. **Confidence Intervals**: Provide uncertainty bounds, not just point estimates
4. **Sensitivity Analysis**: Show which factors most affect results
5. **Empirical Validation**: Compare theoretical bounds to observed transmission rates
6. **Recent Models**: Prioritize Aug/Sep 2025 releases for current relevance

## Advanced Techniques

### Geometric Analysis
- Calculate actual Jacobian of parameter updates
- Measure true effective rank of activation spaces
- Account for attention head redundancy

### Information-Theoretic Rigor
- Apply rate-distortion theory
- Calculate mutual information bounds
- Use PAC learning theory for sample complexity

### Empirical Calibration
- If we had access to trained models, measure:
  - Actual geometric shifts from teacher data
  - Detection thresholds for student models
  - Success rates vs. training budget

## Limitations

- **No Access to Weights**: Estimates based on architecture only, not actual trained models
- **Theoretical Bounds**: Real capacity may be lower due to learning dynamics
- **Same-Model Assumption**: Capacity estimates only valid for teacher-student with same base
- **Linear Assumptions**: May not hold for complex non-linear interactions
- **Noise Model**: SNR estimation based on limited empirical data

## Security & Ethical Considerations

### Research Ethics
- This research is for **understanding**, not enabling malicious transmission
- Focus on defensive capabilities (detecting subliminal signals)
- Transparency about dual-use implications

### Responsible Disclosure
- Share findings with alignment research community
- Don't provide implementation details for exploitation
- Emphasize detection and mitigation strategies

## Integration Notes

### Primitive Sequence
This orchestration uses the following primitive chain:

```
Query (multi-source) → Filter (recent, open-source) → Aggregate (architecture params)
→ Reason (capacity calculation) → Query (validation) → Aggregate (synthesis)
```

### Tool Coordination
- **Sequential dependencies**: Must get paper before calculating
- **Parallel opportunities**: Can search multiple model cards simultaneously
- **Data flow**: Paper insights → Model specs → Calculations → Validation → Report

## Version History

### v1.0.0 (2025-09-29)
- Initial orchestration created
- Covers Llama 3.x, Mistral, Qwen 2.5 models
- Prioritizes Aug-Sep 2025 releases
- Information-theoretic framework established

---

## Development Notes

**Next Enhancements**:
- Add MoE-specific capacity analysis (Mixtral, DeepSeek-V2)
- Include multimodal model architectures
- Develop empirical measurement methodology
- Create visualization tools for capacity vs. model size

**Open Questions**:
- How does quantization affect subliminal channel capacity?
- Can we measure capacity directly without knowing the mechanism?
- Does RLHF training expand or contract available capacity?

---

*For naturalist AI researchers studying emergent phenomena in model-to-model learning*