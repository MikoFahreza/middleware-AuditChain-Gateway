package crypto

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// MerkleProofData menyimpan informasi sibling hash untuk validasi
type MerkleProofData struct {
	SiblingHash string
	TreeLevel   int
}

// MerkleResult menyimpan output akhir agregasi
type MerkleResult struct {
	Root   string
	Proofs map[string][]MerkleProofData // Mapping antara Hash Transaksi dengan jalur Proof-nya
}

// hashNodes menggabungkan dua hash menjadi parent hash menggunakan SHA3-256
func hashNodes(left, right string) string {
	leftBytes, _ := hex.DecodeString(left)
	rightBytes, _ := hex.DecodeString(right)
	combined := append(leftBytes, rightBytes...)

	hash := sha3.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// BuildMerkleTree membangun pohon secara berpasangan dan menghasilkan Root
func BuildMerkleTree(leafHashes []string) *MerkleResult {
	if len(leafHashes) == 0 {
		return nil
	}

	// Inisialisasi map untuk menyimpan proof
	proofs := make(map[string][]MerkleProofData)
	for _, h := range leafHashes {
		proofs[h] = []MerkleProofData{}
	}

	currentLevel := leafHashes
	levelIndex := 0

	// Lakukan perulangan hingga hanya tersisa 1 node (Merkle Root)
	for len(currentLevel) > 1 {
		var nextLevel []string

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right string

			// Jika jumlah node ganjil, node terakhir dipasangkan dengan dirinya sendiri
			if i+1 == len(currentLevel) {
				right = left
			} else {
				right = currentLevel[i+1]
			}

			// Simpan bukti sibling level dasar
			if levelIndex == 0 {
				proofs[left] = append(proofs[left], MerkleProofData{SiblingHash: right, TreeLevel: levelIndex})
				if left != right {
					proofs[right] = append(proofs[right], MerkleProofData{SiblingHash: left, TreeLevel: levelIndex})
				}
			}

			parentHash := hashNodes(left, right)
			nextLevel = append(nextLevel, parentHash)
		}
		currentLevel = nextLevel
		levelIndex++
	}

	return &MerkleResult{
		Root:   currentLevel[0],
		Proofs: proofs,
	}
}

// VerifyMerkleProof merekonstruksi hash dari leaf ke root untuk memverifikasi integritas
func VerifyMerkleProof(transactionHash string, proofs []string, expectedRoot string) bool {
	currentHash := transactionHash

	// Rekonstruksi pohon dengan menggabungkan hash saat ini dengan sibling-nya
	for _, siblingHash := range proofs {
		// Catatan: Dalam skenario ideal, hash diurutkan secara leksikografis (abjad) sebelum digabung
		// agar deterministik tanpa perlu tahu apakah sibling ada di kiri/kanan.
		// Untuk POC ini, menggunakan urutan langsung atau dibalik.

		leftRight := hashNodes(currentHash, siblingHash)
		rightLeft := hashNodes(siblingHash, currentHash)

		// Karena kita tidak menyimpan posisi (Kiri/Kanan) di DB saat ini, kita cek mana yang cocok
		// Pada level berikutnya, currentHash akan menjadi hasil gabungan tersebut.
		// Untuk penyederhanaan verifikasi satu arah di POC:
		currentHash = leftRight

		// Jika ini adalah pengecekan level terakhir dan rightLeft yang benar
		if rightLeft == expectedRoot {
			currentHash = rightLeft
		}
	}

	// Apakah hash yang direkonstruksi sama dengan Merkle Root yang ada di Blockchain/DB?
	return currentHash == expectedRoot
}
