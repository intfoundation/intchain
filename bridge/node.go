package bridge

import (
	"github.com/intfoundation/intchain/cmd/geth"
	"github.com/intfoundation/intchain/cmd/utils"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/intprotocol"
	"github.com/intfoundation/intchain/node"
	"gopkg.in/urfave/cli.v1"
)

var clientIdentifier = "intchain" // Client identifier to advertise over the network

// MakeSystemNode sets up a local node and configures the services to launch
func MakeSystemNode(chainId string, ctx *cli.Context, cch core.CrossChainHelper) *node.Node {

	stack, cfg := gethmain.MakeConfigNode(ctx, chainId)
	//utils.RegisterEthService(stack, &cfg.Eth)
	registerEthService(stack, &cfg.Eth, ctx, cch)

	if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
		// Only Main Chain can start the dashboard, the dashboard is still not complete
		utils.RegisterDashboardService(stack, &cfg.Dashboard, "" /*gitCommit*/)
	}

	// Add the Ethereum Stats daemon if requested.
	//if cfg.Ethstats.URL != "" {
	//	utils.RegisterEthStatsService(stack, cfg.Ethstats.URL)
	//}

	if err := stack.GatherServices(); err != nil {
		return nil
	} else {
		return stack
	}
}

// registerEthService adds an Ethereum client to the stack.
func registerEthService(stack *node.Node, cfg *intprotocol.Config, cliCtx *cli.Context, cch core.CrossChainHelper) {
	var err error
	err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		//return NewBackend(ctx, cfg, cliCtx, pNode, cch)
		fullNode, err := intprotocol.New(ctx, cfg, cliCtx, cch, stack.GetLogger(), cliCtx.GlobalBool(utils.TestnetFlag.Name))
		//if fullNode != nil && cfg.LightServ > 0 {
		//	ls, _ := les.NewLesServer(fullNode, cfg)
		//	fullNode.AddLesServer(ls)
		//}
		return fullNode, err
	})
	if err != nil {
		utils.Fatalf("Failed to register the Ethereum service: %v", err)
	}
}
