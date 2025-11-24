package api

import (
	"testing"

	"github.com/ogen-go/ogen/validate"
)

// Test that Equal() panics when depth limit exceeded
func TestLevel12Equal_DirectPanic(t *testing.T) {
	item := Level12{Value: NewOptString("test")}

	defer func() {
		if r := recover(); r != nil {
			if depthErr, ok := r.(*validate.DepthLimitError); ok {
				t.Logf("Successfully caught panic: %s", depthErr.Error())
			} else {
				t.Fatalf("Expected *validate.DepthLimitError, got %T: %v", r, r)
			}
		} else {
			t.Fatal("Expected panic, but no panic occurred")
		}
	}()

	// Call Equal with depth=11, which should trigger panic (limit is 10)
	item.Equal(item, 11)
}

// Test that fully nested Level1 object triggers panic via Equal()
func TestLevel1Equal_FullyNestedPanic(t *testing.T) {
	item := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							Nested: NewOptLevel7(Level7{
								Nested: NewOptLevel8(Level8{
									Nested: NewOptLevel9(Level9{
										Nested: NewOptLevel10(Level10{
											Nested: NewOptLevel11(Level11{
												Nested: NewOptLevel12(Level12{
													Value: NewOptString("deep-value"),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	}

	defer func() {
		if r := recover(); r != nil {
			if depthErr, ok := r.(*validate.DepthLimitError); ok {
				t.Logf("Successfully caught panic at depth 11: %s", depthErr.Error())
			} else {
				t.Fatalf("Expected *validate.DepthLimitError, got %T: %v", r, r)
			}
		} else {
			t.Fatal("Expected panic from 12-level nesting, but no panic occurred")
		}
	}()

	// Call Equal with depth=0, should recurse to depth=11 and panic
	item.Equal(item, 0)
}

// Test that hashes match for identical items
func TestLevel1Hash_IdenticalItems(t *testing.T) {
	item1 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							Nested: NewOptLevel7(Level7{
								Nested: NewOptLevel8(Level8{
									Nested: NewOptLevel9(Level9{
										Nested: NewOptLevel10(Level10{
											Nested: NewOptLevel11(Level11{
												Nested: NewOptLevel12(Level12{
													Value: NewOptString("deep-value"),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	}

	item2 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							Nested: NewOptLevel7(Level7{
								Nested: NewOptLevel8(Level8{
									Nested: NewOptLevel9(Level9{
										Nested: NewOptLevel10(Level10{
											Nested: NewOptLevel11(Level11{
												Nested: NewOptLevel12(Level12{
													Value: NewOptString("deep-value"),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	}

	hash1 := item1.Hash()
	hash2 := item2.Hash()

	t.Logf("Hash1: %d", hash1)
	t.Logf("Hash2: %d", hash2)

	if hash1 != hash2 {
		t.Fatalf("Hashes should match for identical items, got hash1=%d, hash2=%d", hash1, hash2)
	}
}

// T054: Test depth limit enforcement with 12-level nesting
func TestValidateUniqueLevel1_DepthLimitError(t *testing.T) {
	// Create two identical Level1 objects with all 12 levels nested
	// This should trigger DepthLimitError since depth limit is 10
	item1 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							Nested: NewOptLevel7(Level7{
								Nested: NewOptLevel8(Level8{
									Nested: NewOptLevel9(Level9{
										Nested: NewOptLevel10(Level10{
											Nested: NewOptLevel11(Level11{
												Nested: NewOptLevel12(Level12{
													Value: NewOptString("deep-value"),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	}

	item2 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							Nested: NewOptLevel7(Level7{
								Nested: NewOptLevel8(Level8{
									Nested: NewOptLevel9(Level9{
										Nested: NewOptLevel10(Level10{
											Nested: NewOptLevel11(Level11{
												Nested: NewOptLevel12(Level12{
													Value: NewOptString("deep-value"),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	}

	items := []Level1{item1, item2}

	err := validateUniqueLevel1(items)
	if err == nil {
		t.Fatal("Expected DepthLimitError, got nil")
	}

	depthErr, ok := err.(*validate.DepthLimitError)
	if !ok {
		t.Fatalf("Expected *validate.DepthLimitError, got %T: %v", err, err)
	}

	if depthErr.MaxDepth != 10 {
		t.Errorf("Expected MaxDepth=10, got %d", depthErr.MaxDepth)
	}

	t.Logf("Successfully caught depth limit error: %s", depthErr.Error())
}

// Test that objects within the depth limit work correctly
func TestValidateUniqueLevel1_WithinDepthLimit(t *testing.T) {
	// Create objects with only 5 levels of nesting (well within limit of 10)
	item1 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							// Stop at level 6
						}),
					}),
				}),
			}),
		}),
	}

	item2 := Level1{
		ID: "test-2",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{
							// Stop at level 6
						}),
					}),
				}),
			}),
		}),
	}

	items := []Level1{item1, item2}

	err := validateUniqueLevel1(items)
	if err != nil {
		t.Errorf("Expected no error for items within depth limit, got: %v", err)
	}
}

// Test that duplicates within depth limit are still detected
func TestValidateUniqueLevel1_DuplicateWithinDepthLimit(t *testing.T) {
	// Create two identical objects with 6 levels of nesting
	item1 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{}),
					}),
				}),
			}),
		}),
	}

	item2 := Level1{
		ID: "test-1",
		Nested: NewOptLevel2(Level2{
			Nested: NewOptLevel3(Level3{
				Nested: NewOptLevel4(Level4{
					Nested: NewOptLevel5(Level5{
						Nested: NewOptLevel6(Level6{}),
					}),
				}),
			}),
		}),
	}

	items := []Level1{item1, item2}

	err := validateUniqueLevel1(items)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError, got nil")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T: %v", err, err)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 1 {
		t.Errorf("Expected indices [0, 1], got %v", dupErr.Indices)
	}
}
