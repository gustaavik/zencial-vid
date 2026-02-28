package valueobject

// HashedPassword represents a hashed password.
// The raw password is never stored.
type HashedPassword struct {
	hash string
}

// NewHashedPassword creates a HashedPassword from a hash string.
func NewHashedPassword(hash string) HashedPassword {
	return HashedPassword{hash: hash}
}

// String returns the hash.
func (p HashedPassword) String() string {
	return p.hash
}

// IsZero reports whether the password hash is empty.
func (p HashedPassword) IsZero() bool {
	return p.hash == ""
}
