package gethmain

import (
	"fmt"
	"github.com/intfoundation/intchain/chain"
	"github.com/intfoundation/intchain/consensus/ipbft/consensus"
	"github.com/intfoundation/intchain/internal/debug"
	"github.com/intfoundation/intchain/log"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func intchainCmd(ctx *cli.Context) error {

	if ctx == nil {
		log.Error("oh, ctx is null, how intchain works?")
		return nil
	}

	log.Info("INT Chain is the world's first bottom up new-generation blockchain of things (BoT) communication standard and base application platform.")

	chainMgr := chain.GetCMInstance(ctx)

	// ChildChainFlag flag
	requestChildChain := strings.Split(ctx.GlobalString(ChildChainFlag.Name), ",")

	// Initial P2P Server
	chainMgr.InitP2P()

	// Load Main Chain
	err := chainMgr.LoadMainChain()
	if err != nil {
		log.Errorf("Load Main Chain failed. %v", err)
		return nil
	}

	//set the event.TypeMutex to cch
	chainMgr.InitCrossChainHelper()

	// Start P2P Server
	err = chainMgr.StartP2PServer()
	if err != nil {
		log.Errorf("Start P2P Server failed. %v", err)
		return err
	}
	consensus.NodeID = chainMgr.GetNodeID()[0:16]

	// Start Main Chain
	err = chainMgr.StartMainChain()

	// Load Child Chain
	err = chainMgr.LoadChains(requestChildChain)
	if err != nil {
		log.Errorf("Load Child Chains failed. %v", err)
		return err
	}

	// Start Child Chain
	err = chainMgr.StartChains()
	if err != nil {
		log.Error("start chains failed")
		return err
	}

	err = chainMgr.StartRPC()
	if err != nil {
		log.Error("start intrpc failed")
		return err
	}

	chainMgr.StartInspectEvent()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")

		chainMgr.StopChain()
		chainMgr.WaitChainsStop()
		chainMgr.Stop()

		for i := 3; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Info(fmt.Sprintf("Already shutting down, interrupt %d more times for panic.", i-1))
			}
		}
		debug.Exit() // ensure trace and CPU profile data is flushed.
		debug.LoudPanic("boom")
	}()

	chainMgr.Wait()

	return nil
}
