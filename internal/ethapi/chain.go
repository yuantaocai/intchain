package ethapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/accounts"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/state"
	"github.com/intfoundation/intchain/core/types"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/params"
	"github.com/intfoundation/intchain/rlp"
	"github.com/intfoundation/intchain/rpc"
	"math/big"
	"strings"
	"time"
)

type PublicChainAPI struct {
	am *accounts.Manager
	b  Backend
}

// NewPublicChainAPI creates a new Etheruem protocol API.
func NewPublicChainAPI(b Backend) *PublicChainAPI {
	return &PublicChainAPI{
		am: b.AccountManager(),
		b:  b,
	}
}

func (s *PublicChainAPI) CreateChildChain(ctx context.Context, from common.Address, chainId string,
	minValidators *hexutil.Uint, minDepositAmount *hexutil.Big, startBlock, endBlock *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.CreateChildChain.String(), chainId, uint16(*minValidators), (*big.Int)(minDepositAmount), (*big.Int)(startBlock), (*big.Int)(endBlock))
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.CreateChildChain.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    (*hexutil.Big)(math.MustParseBig256("100000000000000000000000")),
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) JoinChildChain(ctx context.Context, from common.Address, pubkey crypto.BLSPubKey, chainId string,
	depositAmount *hexutil.Big, signature hexutil.Bytes, gasPrice *hexutil.Big) (common.Hash, error) {

	if chainId == "" || strings.Contains(chainId, ";") {
		return common.Hash{}, errors.New("chainId is nil or empty, or contains ';', should be meaningful")
	}

	input, err := intAbi.ChainABI.Pack(intAbi.JoinChildChain.String(), pubkey.Bytes(), chainId, signature)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.JoinChildChain.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    depositAmount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) DepositInMainChain(ctx context.Context, from common.Address, chainId string,
	amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	if chainId == "" || strings.Contains(chainId, ";") {
		return common.Hash{}, errors.New("chainId is nil or empty, or contains ';', should be meaningful")
	}

	if chainId == params.MainnetChainConfig.IntChainId || chainId == params.TestnetChainConfig.IntChainId {
		return common.Hash{}, errors.New("chainId should not be " + params.MainnetChainConfig.IntChainId + " or " + params.TestnetChainConfig.IntChainId)
	}

	input, err := intAbi.ChainABI.Pack(intAbi.DepositInMainChain.String(), chainId)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.DepositInMainChain.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    amount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) DepositInChildChain(ctx context.Context, from common.Address, txHash common.Hash) (common.Hash, error) {

	chainId := s.b.ChainConfig().IntChainId

	input, err := intAbi.ChainABI.Pack(intAbi.DepositInChildChain.String(), chainId, txHash)
	if err != nil {
		return common.Hash{}, err
	}

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      nil,
		GasPrice: nil,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) WithdrawFromChildChain(ctx context.Context, from common.Address,
	amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	chainId := s.b.ChainConfig().IntChainId
	input, err := intAbi.ChainABI.Pack(intAbi.WithdrawFromChildChain.String(), chainId)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.WithdrawFromChildChain.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    amount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) WithdrawFromMainChain(ctx context.Context, from common.Address, amount *hexutil.Big, chainId string, txHash common.Hash) (common.Hash, error) {

	if chainId == params.MainnetChainConfig.IntChainId || chainId == params.TestnetChainConfig.IntChainId {
		return common.Hash{}, errors.New("argument can't be the main chain")
	}

	input, err := intAbi.ChainABI.Pack(intAbi.WithdrawFromMainChain.String(), chainId, (*big.Int)(amount), txHash)
	if err != nil {
		return common.Hash{}, err
	}

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      nil,
		GasPrice: nil,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) GetTxFromChildChainByHash(ctx context.Context, chainId string, txHash common.Hash) (common.Hash, error) {
	cch := s.b.GetCrossChainHelper()

	childTx := cch.GetTX3(chainId, txHash)
	if childTx == nil {
		return common.Hash{}, fmt.Errorf("tx %x does not exist in child chain %s", txHash, chainId)
	}

	return txHash, nil
}

func (s *PublicChainAPI) GetAllTX1(ctx context.Context, from common.Address, blockNr rpc.BlockNumber) ([]common.Hash, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}

	var tx1s []common.Hash
	state.ForEachTX1(from, func(tx1 common.Hash) bool {
		tx1s = append(tx1s, tx1)
		return true
	})
	return tx1s, state.Error()
}

func (s *PublicChainAPI) GetAllTX3(ctx context.Context, from common.Address, blockNr rpc.BlockNumber) ([]common.Hash, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}

	var tx3s []common.Hash
	state.ForEachTX3(from, func(tx3 common.Hash) bool {
		tx3s = append(tx3s, tx3)
		return true
	})
	return tx3s, state.Error()
}

func (s *PublicChainAPI) BroadcastTX3ProofData(ctx context.Context, bs hexutil.Bytes) error {
	chainId := s.b.ChainConfig().IntChainId
	if chainId != params.MainnetChainConfig.IntChainId && chainId != params.TestnetChainConfig.IntChainId {
		return errors.New("this api can only be called in the main chain")
	}

	var proofData types.TX3ProofData
	if err := rlp.DecodeBytes(bs, &proofData); err != nil {
		return err
	}

	cch := s.b.GetCrossChainHelper()
	if err := cch.ValidateTX3ProofData(&proofData); err != nil {
		return err
	}

	// Write to local TX3 cache first.
	cch.WriteTX3ProofData(&proofData)
	// Broadcast the TX3ProofData to all peers in the main chain.
	s.b.BroadcastTX3ProofData(&proofData)
	return nil
}

func (s *PublicChainAPI) GetAllChains() []*ChainStatus {

	cch := s.b.GetCrossChainHelper()
	chainInfoDB := cch.GetChainInfoDB()

	// Load Main Chain
	mainChainId, mainChainEpoch := s.b.GetCrossChainHelper().GetEpochFromMainChain()
	mainChainValidators := make([]*ChainValidator, 0, mainChainEpoch.Validators.Size())
	for _, val := range mainChainEpoch.Validators.Validators {
		mainChainValidators = append(mainChainValidators, &ChainValidator{
			Account:     common.BytesToAddress(val.Address),
			VotingPower: (*hexutil.Big)(val.VotingPower),
		})
	}
	mainChainStatus := &ChainStatus{
		ChainID:    mainChainId,
		Number:     hexutil.Uint64(mainChainEpoch.Number),
		StartTime:  &mainChainEpoch.StartTime,
		Validators: mainChainValidators,
	}

	// Load All Available Child Chain
	chainIds := core.GetChildChainIds(chainInfoDB)

	// Load Complete, now append the data
	result := make([]*ChainStatus, 0, len(chainIds)+1)

	// Add Main Chain Data
	result = append(result, mainChainStatus)

	// Add Child Chain Data
	for _, chainId := range chainIds {
		chainInfo := core.GetChainInfo(chainInfoDB, chainId)

		var chain_status *ChainStatus

		epoch := chainInfo.Epoch
		if epoch == nil {
			chain_status = &ChainStatus{
				ChainID: chainInfo.ChainId,
				Owner:   chainInfo.Owner,
				Message: "child chain not start",
			}
		} else {
			validators := make([]*ChainValidator, 0, epoch.Validators.Size())
			for _, val := range epoch.Validators.Validators {
				validators = append(validators, &ChainValidator{
					Account:     common.BytesToAddress(val.Address),
					VotingPower: (*hexutil.Big)(val.VotingPower),
				})
			}

			chain_status = &ChainStatus{
				ChainID:    chainInfo.ChainId,
				Owner:      chainInfo.Owner,
				Number:     hexutil.Uint64(epoch.Number),
				StartTime:  &epoch.StartTime,
				Validators: validators,
			}
		}
		result = append(result, chain_status)
	}

	return result
}

func (s *PublicChainAPI) SignAddress(from common.Address, consensusPrivateKey hexutil.Bytes) (crypto.Signature, error) {
	if len(consensusPrivateKey) != 32 {
		return nil, errors.New("invalid consensus private key")
	}

	var blsPriv crypto.BLSPrivKey
	copy(blsPriv[:], consensusPrivateKey)

	blsSign := blsPriv.Sign(from.Bytes())

	return blsSign, nil
}

func (s *PublicChainAPI) SetBlockReward(ctx context.Context, from common.Address, reward *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {
	chainId := s.b.ChainConfig().IntChainId
	input, err := intAbi.ChainABI.Pack(intAbi.SetBlockReward.String(), chainId, (*big.Int)(reward))
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.SetBlockReward.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return s.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (s *PublicChainAPI) GetBlockReward(ctx context.Context, blockNr rpc.BlockNumber) (*hexutil.Big, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	return (*hexutil.Big)(state.GetChildChainRewardPerBlock()), nil
}

func (api *PublicChainAPI) WithdrawReward(ctx context.Context, from common.Address, delegateAddress common.Address, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.WithdrawReward.String(), delegateAddress)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.WithdrawReward.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func init() {
	//CreateChildChain
	core.RegisterValidateCb(intAbi.CreateChildChain, ccc_ValidateCb)
	core.RegisterApplyCb(intAbi.CreateChildChain, ccc_ApplyCb)

	//JoinChildChain
	core.RegisterValidateCb(intAbi.JoinChildChain, jcc_ValidateCb)
	core.RegisterApplyCb(intAbi.JoinChildChain, jcc_ApplyCb)

	//DepositInMainChain
	core.RegisterValidateCb(intAbi.DepositInMainChain, dimc_ValidateCb)
	core.RegisterApplyCb(intAbi.DepositInMainChain, dimc_ApplyCb)

	//DepositInChildChain
	core.RegisterValidateCb(intAbi.DepositInChildChain, dicc_ValidateCb)
	core.RegisterApplyCb(intAbi.DepositInChildChain, dicc_ApplyCb)

	//WithdrawFromChildChain
	core.RegisterValidateCb(intAbi.WithdrawFromChildChain, wfcc_ValidateCb)
	core.RegisterApplyCb(intAbi.WithdrawFromChildChain, wfcc_ApplyCb)

	//WithdrawFromMainChain
	core.RegisterValidateCb(intAbi.WithdrawFromMainChain, wfmc_ValidateCb)
	core.RegisterApplyCb(intAbi.WithdrawFromMainChain, wfmc_ApplyCb)

	//SD2MCFuncName
	core.RegisterValidateCb(intAbi.SaveDataToMainChain, sd2mc_ValidateCb)
	core.RegisterApplyCb(intAbi.SaveDataToMainChain, sd2mc_ApplyCb)

	//SetBlockReward
	core.RegisterValidateCb(intAbi.SetBlockReward, sbr_ValidateCb)
	core.RegisterApplyCb(intAbi.SetBlockReward, sbr_ApplyCb)

	// Withdraw reward
	core.RegisterValidateCb(intAbi.WithdrawReward, wdr_ValidateCb)
	core.RegisterApplyCb(intAbi.WithdrawReward, wdr_ApplyCb)
}

func ccc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	chainBalance := state.GetChainBalance(from)
	if chainBalance.Sign() > 0 {
		return errors.New("this address has created one chain, can't create another chain")
	}

	var args intAbi.CreateChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.CreateChildChain.String(), data[4:]); err != nil {
		return err
	}

	if err := cch.CanCreateChildChain(from, args.ChainId, args.MinValidators, args.MinDepositAmount, tx.Value(), args.StartBlock, args.EndBlock); err != nil {
		return err
	}

	return nil
}

func ccc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	chainBalance := state.GetChainBalance(from)
	if chainBalance.Sign() > 0 {
		return errors.New("this address has created one chain, can't create another chain")
	}

	var args intAbi.CreateChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.CreateChildChain.String(), data[4:]); err != nil {
		return err
	}

	startupCost := tx.Value()
	if err := cch.CanCreateChildChain(from, args.ChainId, args.MinValidators, args.MinDepositAmount, startupCost, args.StartBlock, args.EndBlock); err != nil {
		return err
	}

	// Move startup cost from balance to chain balance, it will move to child chain's token pool (address 0x64)
	state.SubBalance(from, startupCost)
	state.AddChainBalance(from, startupCost)

	op := types.CreateChildChainOp{
		From:             from,
		ChainId:          args.ChainId,
		MinValidators:    args.MinValidators,
		MinDepositAmount: args.MinDepositAmount,
		StartBlock:       args.StartBlock,
		EndBlock:         args.EndBlock,
	}
	if ok := ops.Append(&op); !ok {
		return fmt.Errorf("pending ops conflict: %v", op)
	}
	return nil
}

func jcc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.JoinChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.JoinChildChain.String(), data[4:]); err != nil {
		return err
	}

	if err := cch.ValidateJoinChildChain(from, args.PubKey, args.ChainId, tx.Value(), args.Signature); err != nil {
		return err
	}

	return nil
}

func jcc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.JoinChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.JoinChildChain.String(), data[4:]); err != nil {
		return err
	}

	amount := tx.Value()

	if err := cch.ValidateJoinChildChain(from, args.PubKey, args.ChainId, amount, args.Signature); err != nil {
		return err
	}

	var pub crypto.BLSPubKey
	copy(pub[:], args.PubKey)

	op := types.JoinChildChainOp{
		From:          from,
		PubKey:        pub,
		ChainId:       args.ChainId,
		DepositAmount: amount,
	}
	if ok := ops.Append(&op); !ok {
		return fmt.Errorf("pending ops conflict: %v", op)
	}

	// Everything fine, Lock the Balance for this account
	state.SubBalance(from, amount)
	state.AddChildChainDepositBalance(from, args.ChainId, amount)

	return nil
}

func dimc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	var args intAbi.DepositInMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.DepositInMainChain.String(), data[4:]); err != nil {
		return err
	}

	running := core.CheckChildChainRunning(cch.GetChainInfoDB(), args.ChainId)
	if !running {
		return fmt.Errorf("%s chain not running", args.ChainId)
	}

	return nil
}

func dimc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.DepositInMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.DepositInMainChain.String(), data[4:]); err != nil {
		return err
	}

	running := core.CheckChildChainRunning(cch.GetChainInfoDB(), args.ChainId)
	if !running {
		return fmt.Errorf("%s chain not running", args.ChainId)
	}

	// mark from -> tx1 on the main chain (to find all tx1 when given 'from').
	state.AddTX1(from, tx.Hash())

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)

	amount := tx.Value()
	state.SubBalance(from, amount)
	state.AddChainBalance(chainInfo.Owner, amount)

	return nil
}

func dicc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.DepositInChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.DepositInChildChain.String(), data[4:]); err != nil {
		return err
	}

	dimcTx := cch.GetTxFromMainChain(args.TxHash)
	if dimcTx == nil {
		return fmt.Errorf("tx %x does not exist in main chain", args.TxHash)
	}

	if state.HasTX1(from, args.TxHash) {
		return fmt.Errorf("tx %x already used in child chain", args.TxHash)
	}

	signer2 := types.NewEIP155Signer(dimcTx.ChainId())
	dimcFrom, err := types.Sender(signer2, dimcTx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var dimcArgs intAbi.DepositInMainChainArgs
	dimcData := dimcTx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&dimcArgs, intAbi.DepositInMainChain.String(), dimcData[4:]); err != nil {
		return err
	}

	if from != dimcFrom || args.ChainId != dimcArgs.ChainId {
		return errors.New("params are not consistent with tx in main chain")
	}

	return nil
}

func dicc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.DepositInChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.DepositInChildChain.String(), data[4:]); err != nil {
		return err
	}

	dimcTx := cch.GetTxFromMainChain(args.TxHash)
	if dimcTx == nil {
		return fmt.Errorf("tx %x does not exist in main chain", args.TxHash)
	}

	if state.HasTX1(from, args.TxHash) {
		return fmt.Errorf("tx %x already used in child chain", args.TxHash)
	}

	signer2 := types.NewEIP155Signer(dimcTx.ChainId())
	dimcFrom, err := types.Sender(signer2, dimcTx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var dimcArgs intAbi.DepositInMainChainArgs
	dimcData := dimcTx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&dimcArgs, intAbi.DepositInMainChain.String(), dimcData[4:]); err != nil {
		return err
	}

	if from != dimcFrom || args.ChainId != dimcArgs.ChainId {
		return errors.New("params are not consistent with tx in main chain")
	}

	// mark from -> tx1 on the child chain (to indicate tx1's used).
	state.AddTX1(from, args.TxHash)

	state.AddBalance(dimcFrom, dimcTx.Value())

	return nil
}

func wfcc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	var args intAbi.WithdrawFromChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromChildChain.String(), data[4:]); err != nil {
		return err
	}

	return nil
}

func wfcc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.WithdrawFromChildChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromChildChain.String(), data[4:]); err != nil {
		return err
	}

	// mark from -> tx3 on the child chain (to find all tx3 when given 'from').
	state.AddTX3(from, tx.Hash())

	state.SubBalance(from, tx.Value())

	return nil
}

func wfmc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		return err
	}

	if state.HasTX3(from, args.TxHash) {
		return fmt.Errorf("tx %x already used in the main chain", args.TxHash)
	}

	// Notice: there's no validation logic for tx3 here.

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if chainInfo == nil {
		return errors.New("chain id not exist")
	} else if state.GetChainBalance(chainInfo.Owner).Cmp(args.Amount) < 0 {
		return errors.New("no enough balance to withdraw")
	}

	return nil
}

//for tx4 execution, return core.ErrInvalidTx4 if there is error, except need to wait tx3
func wfmc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		//return core.ErrInvalidSender
		return core.ErrInvalidTx4
	}

	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		//return err
		return core.ErrInvalidTx4
	}

	if state.HasTX3(from, args.TxHash) {
		//return fmt.Errorf("tx %x already used in the main chain", args.TxHash)
		return core.ErrInvalidTx4
	}

	if mining { // validate only when mining.
		wfccTx := cch.GetTX3(args.ChainId, args.TxHash)
		if wfccTx == nil {
			return fmt.Errorf("tx %x does not exist in child chain %s", args.TxHash, args.ChainId)
		}

		signer2 := types.NewEIP155Signer(wfccTx.ChainId())
		wfccFrom, err := types.Sender(signer2, wfccTx)
		if err != nil {
			//return core.ErrInvalidSender
			return core.ErrInvalidTx4
		}

		var wfccArgs intAbi.WithdrawFromChildChainArgs
		wfccData := wfccTx.Data()
		if err := intAbi.ChainABI.UnpackMethodInputs(&wfccArgs, intAbi.WithdrawFromChildChain.String(), wfccData[4:]); err != nil {
			//return err
			return core.ErrInvalidTx4
		}

		if from != wfccFrom || args.ChainId != wfccArgs.ChainId || args.Amount.Cmp(wfccTx.Value()) != 0 {
			return core.ErrInvalidTx4
		}
	}

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if state.GetChainBalance(chainInfo.Owner).Cmp(args.Amount) < 0 {
		//return errors.New("no enough balance to withdraw")
		return core.ErrInvalidTx4
	}

	// mark from -> tx3 on the main chain (to indicate tx3's used).
	state.AddTX3(from, args.TxHash)

	state.SubChainBalance(chainInfo.Owner, args.Amount)
	state.AddBalance(from, args.Amount)

	return nil
}

func sd2mc_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	var bs []byte
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&bs, intAbi.SaveDataToMainChain.String(), data[4:]); err != nil {
		return err
	}

	err := cch.VerifyChildChainProofData(bs)
	if err != nil {
		return fmt.Errorf("data can not pass verification: %v", err)
	}

	return nil
}

func sd2mc_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {
	var bs []byte
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&bs, intAbi.SaveDataToMainChain.String(), data[4:]); err != nil {
		return err
	}

	// Validate only when mining
	if mining {
		err := cch.VerifyChildChainProofData(bs)
		if err != nil {
			return fmt.Errorf("data can not pass verification: %v", err)
		}
	}

	op := types.SaveDataToMainChainOp{
		Data: bs,
	}
	if ok := ops.Append(&op); !ok {
		return fmt.Errorf("pending ops conflict: %v", op)
	}

	return nil
}

func sbr_ValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {
	from := derivedAddressFromTx(tx)
	_, verror := setBlockRewardValidation(from, tx, cch)
	if verror != nil {
		return verror
	}
	return nil
}

func sbr_ApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {
	from := derivedAddressFromTx(tx)
	args, verror := setBlockRewardValidation(from, tx, cch)
	if verror != nil {
		return verror
	}

	state.SetChildChainRewardPerBlock(args.Reward)
	return nil
}

func wdr_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, err := withDrawRewardValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func wdr_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	from := derivedAddressFromTx(tx)

	args, err := withDrawRewardValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	reward := state.GetRewardBalanceByDelegateAddress(from, args.DelegateAddress)
	state.SubRewardBalanceByDelegateAddress(from, args.DelegateAddress, reward)
	state.AddBalance(from, reward)

	return nil
}

type ChainStatus struct {
	ChainID    string            `json:"chain_id"`
	Owner      common.Address    `json:"owner"`
	Number     hexutil.Uint64    `json:"current_epoch,omitempty"`
	StartTime  *time.Time        `json:"epoch_start_time,omitempty"`
	Validators []*ChainValidator `json:"validators,omitempty"`

	Message string `json:"message,omitempty"`
}

type ChainValidator struct {
	Account     common.Address `json:"address"`
	VotingPower *hexutil.Big   `json:"voting_power"`
}

// Validation

func setBlockRewardValidation(from common.Address, tx *types.Transaction, cch core.CrossChainHelper) (*intAbi.SetBlockRewardArgs, error) {

	var args intAbi.SetBlockRewardArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.SetBlockReward.String(), data[4:]); err != nil {
		return nil, err
	}

	ci := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if ci == nil || ci.Owner != from {
		return nil, core.ErrNotOwner
	}

	if args.Reward.Sign() == -1 {
		return nil, core.ErrNegativeValue
	}

	return &args, nil
}

func withDrawRewardValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.WithdrawRewardArgs, error) {

	var args intAbi.WithdrawRewardArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawReward.String(), data[4:]); err != nil {
		return nil, err
	}

	reward := state.GetRewardBalanceByDelegateAddress(from, args.DelegateAddress)

	if reward.Sign() < 1 {
		return nil, fmt.Errorf("have no reward to withdraw")
	}

	//if args.Amount.Cmp(reward) == 1 {
	//	return nil, fmt.Errorf("reward balance not enough, withdraw amount %v, but balance %v, delegate address %v", args.Amount, reward, args.DelegateAddress)
	//}
	return &args, nil
}

func wfmcValidateCb(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		return err
	}

	if state.HasTX3(from, args.TxHash) {
		return fmt.Errorf("tx %x already used in the main chain", args.TxHash)
	}

	// Notice: there's no validation logic for tx3 here.

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if chainInfo == nil {
		return errors.New("chain id not exist")
	} else if state.GetChainBalance(chainInfo.Owner).Cmp(args.Amount) < 0 {
		return errors.New("no enough balance to withdraw")
	}

	return nil
}

//for tx4 execution, return core.ErrInvalidTx4 if there is error, except need to wait tx3
func wfmcApplyCb(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper, mining bool) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidTx4
		//return core.ErrInvalidSender
	}

	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		return core.ErrInvalidTx4
		//return err
	}

	if state.HasTX3(from, args.TxHash) {
		return core.ErrInvalidTx4
		//return fmt.Errorf("tx %x already used in the main chain", args.TxHash)
	}

	if mining { // validate only when mining.
		wfccTx := cch.GetTX3(args.ChainId, args.TxHash)
		if wfccTx == nil {
			return fmt.Errorf("tx %x does not exist in child chain %s", args.TxHash, args.ChainId)
		}

		signer2 := types.NewEIP155Signer(wfccTx.ChainId())
		wfccFrom, err := types.Sender(signer2, wfccTx)
		if err != nil {
			return core.ErrInvalidTx4
			//return core.ErrInvalidSender
		}

		var wfccArgs intAbi.WithdrawFromChildChainArgs
		wfccData := wfccTx.Data()
		if err := intAbi.ChainABI.UnpackMethodInputs(&wfccArgs, intAbi.WithdrawFromChildChain.String(), wfccData[4:]); err != nil {
			return core.ErrInvalidTx4
			//return err
		}

		if from != wfccFrom || args.ChainId != wfccArgs.ChainId || args.Amount.Cmp(wfccTx.Value()) != 0 {
			return core.ErrInvalidTx4
		}
	}

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if state.GetChainBalance(chainInfo.Owner).Cmp(args.Amount) < 0 {
		return core.ErrInvalidTx4
		//return errors.New("no enough balance to withdraw")
	}

	// mark from -> tx3 on the main chain (to indicate tx3's used).
	state.AddTX3(from, args.TxHash)

	state.SubChainBalance(chainInfo.Owner, args.Amount)
	state.AddBalance(from, args.Amount)

	return nil
}

func wfmcValidateCbV1(tx *types.Transaction, state *state.StateDB, cch core.CrossChainHelper) error {

	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return core.ErrInvalidSender
	}

	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		return err
	}

	if state.HasTX3(from, args.TxHash) {
		return fmt.Errorf("tx %x already used in the main chain", args.TxHash)
	}

	// Notice: there's validation logic for tx3 here.
	{
		wfccTx := cch.GetTX3(args.ChainId, args.TxHash)
		if wfccTx == nil {
			return fmt.Errorf("tx %x does not exist in child chain %s", args.TxHash, args.ChainId)
		}

		signer2 := types.NewEIP155Signer(wfccTx.ChainId())
		wfccFrom, err := types.Sender(signer2, wfccTx)
		if err != nil {
			return core.ErrInvalidSender
		}

		var wfccArgs intAbi.WithdrawFromChildChainArgs
		wfccData := wfccTx.Data()
		if err := intAbi.ChainABI.UnpackMethodInputs(&wfccArgs, intAbi.WithdrawFromChildChain.String(), wfccData[4:]); err != nil {
			return err
		}

		if from != wfccFrom || args.ChainId != wfccArgs.ChainId || args.Amount.Cmp(wfccTx.Value()) != 0 {
			return core.ErrInvalidTx4
		}
	}

	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)
	if chainInfo == nil {
		return errors.New("chain id not exist")
	} else if state.GetChainBalance(chainInfo.Owner).Cmp(args.Amount) < 0 {
		return errors.New("no enough balance to withdraw")
	}

	return nil
}

//for tx4 execution, return core.ErrInvalidTx4 if there is error, except need to wait tx3
func wfmcApplyCbV1(tx *types.Transaction, state *state.StateDB, ops *types.PendingOps, cch core.CrossChainHelper) error {

	if err := wfmcValidateCbV1(tx, state, cch); err != nil {
		return err
	}

	from, _ := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	var args intAbi.WithdrawFromMainChainArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawFromMainChain.String(), data[4:]); err != nil {
		return err
	}
	chainInfo := core.GetChainInfo(cch.GetChainInfoDB(), args.ChainId)

	// mark from -> tx3 on the main chain (to indicate tx3's used).
	state.AddTX3(from, args.TxHash)

	state.SubChainBalance(chainInfo.Owner, args.Amount)
	state.AddBalance(from, args.Amount)

	return nil
}
