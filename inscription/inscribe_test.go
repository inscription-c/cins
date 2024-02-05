package inscription

import (
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/signal"
	"math"
	"math/rand"
	"testing"
	"time"
)

var (
	dbAddr       = "127.0.0.1:3306"
	dbUser       = "root"
	dbPass       = "root"
	databaseName = "cins"
)

func TestInscribe(t *testing.T) {
	testnet = true
	postage = 1
	inscriptionsFilePath = "./test/cbrc20.json"
	unlockConditionFile = "./test/unlock_condition.json"
	destination = "tb1qq2lsrdnylv0qu7eezsruhv29jxrujm3fpzfpkf"
	if err := inscribe(); err != nil {
		t.Fatal(err)
	}
	<-signal.InterruptHandlersDone
}

func TestDeleteMockInscriptions(t *testing.T) {
	db, err := dao.NewDB(
		dao.WithAddr(dbAddr),
		dao.WithUser(dbUser),
		dao.WithPassword(dbPass),
		dao.WithDBName(databaseName),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *dao.DB) error {
		if err := tx.DeleteMockInscriptions(); err != nil {
			return err
		}
		if err := tx.DeleteMockProtocol(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMockInscriptions(t *testing.T) {
	db, err := dao.NewDB(
		dao.WithAddr(dbAddr),
		dao.WithUser(dbUser),
		dao.WithPassword(dbPass),
		dao.WithDBName(databaseName),
	)
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]string{
		"p":    "c-brc-20",
		"op":   "deploy",
		"tick": "cins",
		"max":  "21000000",
		"lim":  "1000",
	}

	if err := db.Transaction(func(tx *dao.DB) error {
		for i := int64(1); i <= 100; i++ {
			body["tick"] = fmt.Sprintf("cins-%d", i)
			bodyBs := gconv.Bytes(body)
			insId := tables.InscriptionId{
				TxId: fmt.Sprintf("%064d", i),
			}
			owner := "tb1qq2lsrdnylv0qu7eezsruhv29jxrujm3fpzfpkf"
			if err := tx.CreateInscription(&tables.Inscriptions{
				InscriptionId:   insId,
				Index:           0,
				SequenceNum:     -i,
				InscriptionNum:  math.MaxInt64 - i,
				Owner:           owner,
				Charms:          0,
				Fee:             uint64(rand.Intn(10000)),
				Height:          uint32(i),
				Sat:             0,
				Timestamp:       time.Now().Unix(),
				Body:            bodyBs,
				ContentEncoding: "",
				ContentType:     constants.ContentTypeJson.String(),
				MediaType:       constants.ContentTypeJson.MediaType().String(),
				ContentSize:     uint32(len(bodyBs)),
				UnlockCondition: tables.UnlockCondition{
					Type:     constants.UnlockConditionTypeBlockchain,
					Chain:    "309",
					Contract: "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqupnqt6y5nu39j0704jvw770esjfdzulzsyqwqes9az2f7gje8l86ex8008ucfyk3w03gk2pfrr",
				},
				Metadata: nil,
				Pointer:  0,
			}); err != nil {
				return err
			}

			if err := tx.SaveProtocol(&tables.Protocol{
				InscriptionId: insId,
				Index:         0,
				SequenceNum:   -i,
				Protocol:      body["p"],
				Ticker:        body["tick"],
				Operator:      body["op"],
				Owner:         owner,
				Max:           gconv.Uint64(body["max"]),
				Limit:         gconv.Uint64(body["lim"]),
				Decimals:      18,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
