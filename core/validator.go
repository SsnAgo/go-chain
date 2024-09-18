package core

type BlockValidator struct {
	bc *Blockchain
}

func  (v *BlockValidator) Validate(o any) bool {
	b, ok := o.(Block)
	if !ok {
		v.bc.logger.Printf("Invalid block data: %v", o)
		return false
	}
	if v.bc.HasBlock(b.GetDataHash()) {
		v.bc.logger.Printf("Block %s already exists", b.GetDataHash())
		return false
	}
	if b.Height() != v.bc.Height()+1 {
		v.bc.logger.Printf("Invalid block height: %d, current height: %d", b.Height(), v.bc.Height())
		return false
	}
	currLb := v.bc.GetLatestBlock()
	if currLb.GetDataHash() != b.GetPrevBlockHash() {
		v.bc.logger.Printf("Invalid block prev hash: %s, current hash: %s", b.GetPrevBlockHash(), currLb.GetDataHash())
		return false
	}
	if !b.Verify() {
		v.bc.logger.Printf("Invalid block: %+v", b)
		return false
	}
	return true
}