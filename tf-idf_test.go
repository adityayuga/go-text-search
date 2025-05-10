package gotextsearch

import (
	"testing"
)

func Test_invertedIndex(t *testing.T) {
	idx := New()
	idx.SeedDocuments([]string{
		"Hello world",
		"Ameren (St. Louis)",
		"Andy's Frozen Custard (Springfield)",
		"Anheuser-Busch (St. Louis)",
		"Arch Resources (St. Louis)",
		"B&B Theatres (Liberty)",
		"Baron Aviation Services (Vichy)",
		"Bass Pro Shops (Springfield)",
		"Caleres (Clayton)",
		"Centene Corporation (St. Louis)",
		"Columbia Insurance Group (Columbia)",
		"Commerce Bancshares (Kansas City)",
		"Dierbergs Markets (Chesterfield)",
		"Drury Hotels (Creve Coeur)",
		"Edward Jones Investments (St. Louis)",
		"Emerson Electric (Ferguson)",
		"Energizer (Town and Country)",
		"Enterprise Rent-A-Car (Clayton)",
		"Express Scripts (St. Louis County)",
		"Ferrellgas (Liberty)",
		"GoJet Airlines (Bridgeton)",
		"Graybar (Clayton)",
		"H&R Block (Kansas City)",
		"PT Bank Mandiri (Persero) Tbk (Jakarta)",
	})

	results := idx.Search("Hello", 10)
	if results[0].Text != "hello world" {
		t.Errorf("Search() = %v, want %v", results, "hello world")
	}

	results2 := idx.Search("corpration", 10)
	if results2[0].Text != "centene corporation (st. louis)" {
		t.Errorf("Search() = %v, want %v", results2, "centene corporation (st. louis)")
	}

	results3 := idx.Search("mandiri", 10)
	if results3[0].Text != "pt bank mandiri (persero) tbk (jakarta)" {
		t.Errorf("Search() = %v, want %v", results3, "PT Bank Mandiri (Persero) Tbk (Jakarta)")
	}
}
