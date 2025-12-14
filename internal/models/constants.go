package models

// ConfidenceThreshold is the minimum confidence score for database results
// before falling back to LLM queries
const ConfidenceThreshold = 0.8

// LLMConfidenceScore is the confidence score assigned to LLM-generated results
const LLMConfidenceScore = 0.5

// EmbeddingDimension is the dimension of the OpenAI text-embedding-3-small model
const EmbeddingDimension = 768
