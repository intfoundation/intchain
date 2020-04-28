package main

import (
	"fmt"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/core/rawdb"
	"github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/log"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	"encoding/json"
	cmn "github.com/intfoundation/go-common"
	cfg "github.com/intfoundation/go-config"
	dbm "github.com/intfoundation/go-db"
	"github.com/intfoundation/intchain/accounts/keystore"
	"github.com/intfoundation/intchain/cmd/utils"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/params"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	POSReward = "200000000000000000000000000" // 2亿
	//LockReward = "11500000000000000000000000"  // 11.5m

	TotalYear = 10

	DefaultAccountPassword = "intchain"
)

type BalaceAmount struct {
	balance string
	amount  string
}

type InvalidArgs struct {
	args string
}

func (invalid InvalidArgs) Error() string {
	return "invalid args:" + invalid.args
}

func initIntGenesis(ctx *cli.Context) error {
	log.Info("this is init_int_genesis")
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("len of args is %d", len(args))
		return nil
	}
	balance_str := args[0]

	chainId := MainChain
	isMainnet := true
	if ctx.GlobalBool(utils.TestnetFlag.Name) {
		chainId = TestnetChain
		isMainnet = false
	}
	log.Infof("this is init_int_genesis chainId %v", chainId)
	log.Info("this is init_int_genesis" + ctx.GlobalString(utils.DataDirFlag.Name) + "--" + ctx.Args()[0])
	return init_int_genesis(utils.GetTendermintConfig(chainId, ctx), balance_str, isMainnet)
}

func init_int_genesis(config cfg.Config, balanceStr string, isMainnet bool) error {

	balanceAmounts, err := parseBalaceAmount(balanceStr)
	if err != nil {
		utils.Fatalf("init int_genesis_file failed")
		return err
	}

	validators := createPriValidators(config, len(balanceAmounts))
	extraData, _ := hexutil.Decode("0x0")

	var chainConfig *params.ChainConfig
	if isMainnet {
		chainConfig = params.MainnetChainConfig
	} else {
		chainConfig = params.TestnetChainConfig
	}

	var coreGenesis = core.GenesisWrite{
		Config:     chainConfig,
		Nonce:      0xdeadbeefdeadbeef,
		Timestamp:  uint64(time.Now().Unix()),
		ParentHash: common.Hash{},
		ExtraData:  extraData,
		GasLimit:   0x4c4b400,
		Difficulty: new(big.Int).SetUint64(0x01),
		Mixhash:    common.Hash{},
		Coinbase:   "INT3AAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Alloc:      core.GenesisAllocWrite{},
	}
	for i, validator := range validators {
		coreGenesis.Alloc[validator.Address.String()] = core.GenesisAccount{
			Balance: math.MustParseBig256(balanceAmounts[i].balance),
			Amount:  math.MustParseBig256(balanceAmounts[i].amount),
		}
	}

	contents, err := json.MarshalIndent(coreGenesis, "", "\t")
	if err != nil {
		utils.Fatalf("marshal coreGenesis failed")
		return err
	}
	intGenesisPath := config.GetString("int_genesis_file")

	if err = ioutil.WriteFile(intGenesisPath, contents, 0654); err != nil {
		utils.Fatalf("write int_genesis_file failed")
		return err
	}
	return nil
}

func initCmd(ctx *cli.Context) error {

	// int genesis.json
	intGenesisPath := ctx.Args().First()
	fmt.Printf("int genesis path %v\n", intGenesisPath)
	if len(intGenesisPath) == 0 {
		utils.Fatalf("must supply path to genesis JSON file")
	}

	chainId := ctx.Args().Get(1)
	if chainId == "" {
		chainId = MainChain
		if ctx.GlobalBool(utils.TestnetFlag.Name) {
			chainId = TestnetChain
		}
	}

	return init_cmd(ctx, utils.GetTendermintConfig(chainId, ctx), chainId, intGenesisPath)
}

func InitChildChainCmd(ctx *cli.Context) error {
	// Load ChainInfo db
	chainInfoDb := dbm.NewDB("chaininfo", "leveldb", ctx.GlobalString(utils.DataDirFlag.Name))
	if chainInfoDb == nil {
		return errors.New("could not open chain info database")
	}
	defer chainInfoDb.Close()

	// Initial Child Chain Genesis
	childChainIds := ctx.GlobalString("childChain")
	if childChainIds == "" {
		return errors.New("please provide child chain id to initialization")
	}

	chainIds := strings.Split(childChainIds, ",")
	for _, chainId := range chainIds {
		ethGenesis, tdmGenesis := core.LoadChainGenesis(chainInfoDb, chainId)
		if ethGenesis == nil || tdmGenesis == nil {
			return errors.New(fmt.Sprintf("unable to retrieve the genesis file for child chain %s", chainId))
		}

		childConfig := utils.GetTendermintConfig(chainId, ctx)

		// Write down genesis and get the genesis path
		ethGenesisPath := childConfig.GetString("int_genesis_file")
		if err := ioutil.WriteFile(ethGenesisPath, ethGenesis, 0644); err != nil {
			utils.Fatalf("write int_genesis_file failed")
			return err
		}

		// Init the blockchain from genesis path
		init_int_blockchain(chainId, ethGenesisPath, ctx)

		// Write down TDM Genesis directly
		if err := ioutil.WriteFile(childConfig.GetString("genesis_file"), tdmGenesis, 0644); err != nil {
			utils.Fatalf("write tdm genesis_file failed")
			return err
		}

	}

	return nil
}

func init_cmd(ctx *cli.Context, config cfg.Config, chainId string, intGenesisPath string) error {

	init_int_blockchain(chainId, intGenesisPath, ctx)

	init_em_files(config, chainId, intGenesisPath, nil)

	return nil
}

func init_int_blockchain(chainId string, intGenesisPath string, ctx *cli.Context) {

	dbPath := filepath.Join(utils.MakeDataDir(ctx), chainId, "int/chaindata")
	log.Infof("init_int_blockchain 0 with dbPath: %s", dbPath)

	chainDb, err := rawdb.NewLevelDBDatabase(filepath.Join(utils.MakeDataDir(ctx), chainId, clientIdentifier, "chaindata"), 0, 0, "int/db/chaindata/")
	if err != nil {
		utils.Fatalf("could not open database: %v", err)
	}
	defer chainDb.Close()

	log.Info("init_int_blockchain 1")
	genesisFile, err := os.Open(intGenesisPath)
	if err != nil {
		utils.Fatalf("failed to read genesis file: %v", err)
	}
	defer genesisFile.Close()

	log.Info("init_int_blockchain 2")
	block, err := core.WriteGenesisBlock(chainDb, genesisFile)
	if err != nil {
		utils.Fatalf("failed to write genesis block: %v", err)
	}

	log.Info("init_int_blockchain end")
	log.Infof("successfully wrote genesis block and/or chain rule set: %x", block.Hash())
}

func init_em_files(config cfg.Config, chainId string, genesisPath string, validators []types.GenesisValidator) error {
	gensisFile, err := os.Open(genesisPath)
	defer gensisFile.Close()
	if err != nil {
		utils.Fatalf("failed to read intchain genesis file: %v", err)
		return err
	}
	contents, err := ioutil.ReadAll(gensisFile)
	if err != nil {
		utils.Fatalf("failed to read intchain genesis file: %v", err)
		return err
	}
	var (
		genesisW    core.GenesisWrite
		coreGenesis core.Genesis
	)
	if err := json.Unmarshal(contents, &genesisW); err != nil {
		return err
	}

	coreGenesis = core.Genesis{
		Config:     genesisW.Config,
		Nonce:      genesisW.Nonce,
		Timestamp:  genesisW.Timestamp,
		ParentHash: genesisW.ParentHash,
		ExtraData:  genesisW.ExtraData,
		GasLimit:   genesisW.GasLimit,
		Difficulty: genesisW.Difficulty,
		Mixhash:    genesisW.Mixhash,
		Coinbase:   common.StringToAddress(genesisW.Coinbase),
		Alloc:      core.GenesisAlloc{},
	}

	for k, v := range genesisW.Alloc {
		coreGenesis.Alloc[common.StringToAddress(k)] = v
	}

	var privValidator *types.PrivValidator
	// validators == nil means we are init the Genesis from priv_validator, not from runtime GenesisValidator
	if validators == nil {
		privValPath := config.GetString("priv_validator_file")
		if _, err := os.Stat(privValPath); os.IsNotExist(err) {
			log.Info("priv_validator_file not exist, probably you are running in non-mining mode")
			return nil
		}
		// Now load the priv_validator_file
		privValidator = types.LoadPrivValidator(privValPath)
	}

	// Create the Genesis Doc
	if err := createGenesisDoc(config, chainId, &coreGenesis, privValidator, validators); err != nil {
		utils.Fatalf("failed to write genesis file: %v", err)
		return err
	}
	return nil
}

func createGenesisDoc(config cfg.Config, chainId string, coreGenesis *core.Genesis, privValidator *types.PrivValidator, validators []types.GenesisValidator) error {
	genFile := config.GetString("genesis_file")
	if _, err := os.Stat(genFile); os.IsNotExist(err) {

		posReward, _ := new(big.Int).SetString(POSReward, 10)
		totalYear := TotalYear
		rewardFirstYear := new(big.Int).Div(posReward, big.NewInt(int64(totalYear)))

		var rewardScheme types.RewardSchemeDoc
		if chainId == MainChain || chainId == TestnetChain {
			rewardScheme = types.RewardSchemeDoc{
				TotalReward:        posReward,
				RewardFirstYear:    rewardFirstYear,
				EpochNumberPerYear: 4380,
				TotalYear:          uint64(totalYear),
			}
		} else {
			rewardScheme = types.RewardSchemeDoc{
				TotalReward:        big.NewInt(0),
				RewardFirstYear:    big.NewInt(0),
				EpochNumberPerYear: 12,
				TotalYear:          0,
			}
		}

		var rewardPerBlock *big.Int
		if chainId == MainChain || chainId == TestnetChain {
			rewardPerBlock = big.NewInt(634195839675291700)
		} else {
			rewardPerBlock = big.NewInt(0)
		}

		fmt.Printf("init reward block %v\n", rewardPerBlock)
		genDoc := types.GenesisDoc{
			ChainID:      chainId,
			Consensus:    types.CONSENSUS_IPBFT,
			GenesisTime:  time.Now(),
			RewardScheme: rewardScheme,
			CurrentEpoch: types.OneEpochDoc{
				Number:         0,
				RewardPerBlock: rewardPerBlock,
				StartBlock:     0,
				EndBlock:       7200,
				Status:         0,
			},
		}

		if privValidator != nil {
			coinbase, amount, checkErr := checkAccount(*coreGenesis)
			if checkErr != nil {
				log.Infof(checkErr.Error())
				cmn.Exit(checkErr.Error())
			}

			genDoc.CurrentEpoch.Validators = []types.GenesisValidator{{
				EthAccount: coinbase,
				PubKey:     privValidator.PubKey,
				Amount:     amount,
			}}
		} else if validators != nil {
			genDoc.CurrentEpoch.Validators = validators
		}
		genDoc.SaveAs(genFile)
	}
	return nil
}

func generateTDMGenesis(childChainID string, validators []types.GenesisValidator) ([]byte, error) {
	var rewardScheme = types.RewardSchemeDoc{
		TotalReward:        big.NewInt(0),
		RewardFirstYear:    big.NewInt(0),
		EpochNumberPerYear: 12,
		TotalYear:          0,
	}

	genDoc := types.GenesisDoc{
		ChainID:      childChainID,
		Consensus:    types.CONSENSUS_IPBFT,
		GenesisTime:  time.Now(),
		RewardScheme: rewardScheme,
		CurrentEpoch: types.OneEpochDoc{
			Number:         0,
			RewardPerBlock: big.NewInt(0),
			StartBlock:     0,
			EndBlock:       657000,
			Status:         0,
			Validators:     validators,
		},
	}

	contents, err := json.Marshal(genDoc)
	if err != nil {
		utils.Fatalf("marshal tdm Genesis failed")
		return nil, err
	}
	return contents, nil
}

func parseBalaceAmount(s string) ([]*BalaceAmount, error) {
	r, _ := regexp.Compile("\\{[\\ \\t]*\\d+(\\.\\d+)?[\\ \\t]*\\,[\\ \\t]*\\d+(\\.\\d+)?[\\ \\t]*\\}")
	parse_strs := r.FindAllString(s, -1)
	if len(parse_strs) == 0 {
		return nil, InvalidArgs{s}
	}
	balanceAmounts := make([]*BalaceAmount, len(parse_strs))
	for i, v := range parse_strs {
		length := len(v)
		balanceAmount := strings.Split(v[1:length-1], ",")
		if len(balanceAmount) != 2 {
			return nil, InvalidArgs{s}
		}
		balanceAmounts[i] = &BalaceAmount{strings.TrimSpace(balanceAmount[0]), strings.TrimSpace(balanceAmount[1])}
	}
	return balanceAmounts, nil
}

func createPriValidators(config cfg.Config, num int) []*types.PrivValidator {
	validators := make([]*types.PrivValidator, num)

	ks := keystore.NewKeyStore(config.GetString("keystore"), keystore.StandardScryptN, keystore.StandardScryptP)

	privValFile := config.GetString("priv_validator_file_root")
	for i := 0; i < num; i++ {
		// Create New IntChain Account
		account, err := ks.NewAccount(DefaultAccountPassword)
		if err != nil {
			utils.Fatalf("Failed to create IntChain account: %v", err)
		}
		// Generate Consensus KeyPair
		validators[i] = types.GenPrivValidatorKey(account.Address)
		log.Info("createPriValidators", "account:", validators[i].Address, "pwd:", DefaultAccountPassword)
		if i > 0 {
			validators[i].SetFile(privValFile + strconv.Itoa(i) + ".json")
		} else {
			validators[i].SetFile(privValFile + ".json")
		}
		validators[i].Save()
	}
	return validators
}

func checkAccount(coreGenesis core.Genesis) (common.Address, *big.Int, error) {

	coinbase := coreGenesis.Coinbase
	log.Infof("checkAccount(), coinbase is %v", coinbase.String())

	var act common.Address
	amount := big.NewInt(-1)
	balance := big.NewInt(-1)
	found := false
	for address, account := range coreGenesis.Alloc {
		log.Infof("checkAccount(), address is %v, balance is %v, amount is %v", address.String(), account.Balance, account.Amount)
		balance = account.Balance
		amount = account.Amount
		act = address
		found = true
		break
	}

	if !found {
		log.Error("invalidate eth_account")
		return common.Address{}, nil, errors.New("invalidate eth_account")
	}

	if balance.Sign() == -1 || amount.Sign() == -1 {
		log.Errorf("balance / amount can't be negative integer, balance is %v, amount is %v", balance, amount)
		return common.Address{}, nil, errors.New("no enough balance")
	}

	return act, amount, nil
}

func initEthGenesisFromExistValidator(childChainID string, childConfig cfg.Config, validators []types.GenesisValidator) error {

	contents, err := generateETHGenesis(childChainID, validators)
	if err != nil {
		return err
	}
	ethGenesisPath := childConfig.GetString("int_genesis_file")
	if err = ioutil.WriteFile(ethGenesisPath, contents, 0654); err != nil {
		utils.Fatalf("write int_genesis_file failed")
		return err
	}
	return nil
}

func generateETHGenesis(childChainID string, validators []types.GenesisValidator) ([]byte, error) {
	var coreGenesis = core.Genesis{
		Config:     params.NewChildChainConfig(childChainID),
		Nonce:      0xdeadbeefdeadbeef,
		Timestamp:  0x0,
		ParentHash: common.Hash{},
		ExtraData:  []byte("0x0"),
		GasLimit:   0x8000000,
		Difficulty: new(big.Int).SetUint64(0x400),
		Mixhash:    common.Hash{},
		Coinbase:   common.Address{},
		Alloc:      core.GenesisAlloc{},
	}
	for _, validator := range validators {
		coreGenesis.Alloc[validator.EthAccount] = core.GenesisAccount{
			Balance: big.NewInt(0),
			Amount:  validator.Amount,
		}
	}

	// Add Child Chain Default Token
	coreGenesis.Alloc[abi.ChildChainTokenIncentiveAddr] = core.GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e+18)),
		Amount:  common.Big0,
	}

	contents, err := json.Marshal(coreGenesis)
	if err != nil {
		utils.Fatalf("marshal coreGenesis failed")
		return nil, err
	}
	return contents, nil
}
