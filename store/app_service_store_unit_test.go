package store_test

import (
	"github.com/cloudfoundry/storeadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/loggregatorlib/appservice"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
)

type FakeAdapter struct {
	DeleteCount int
}

func (adapter *FakeAdapter) Connect() error                      { return nil }
func (adapter *FakeAdapter) Create(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) Update(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) CompareAndSwap(storeadapter.StoreNode, storeadapter.StoreNode) error {
	return nil
}
func (adapter *FakeAdapter) CompareAndSwapByIndex(uint64, storeadapter.StoreNode) error {
	return nil
}

func (adapter *FakeAdapter) SetMulti(nodes []storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) Get(key string) (storeadapter.StoreNode, error) {
	return storeadapter.StoreNode{}, nil
}
func (adapter *FakeAdapter) ListRecursively(key string) (storeadapter.StoreNode, error) {
	return storeadapter.StoreNode{}, nil
}
func (adapter *FakeAdapter) Delete(keys ...string) error {
	adapter.DeleteCount++
	return nil
}
func (adapter *FakeAdapter) CompareAndDelete(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) UpdateDirTTL(key string, ttl uint64) error     { return nil }
func (adapter *FakeAdapter) Watch(key string) (events <-chan storeadapter.WatchEvent, stop chan<- bool, errors <-chan error) {
	return nil, nil, nil
}
func (adapter *FakeAdapter) Disconnect() error { return nil }
func (adapter *FakeAdapter) MaintainNode(storeNode storeadapter.StoreNode) (lostNode <-chan bool, releaseNode chan chan bool, err error) {
	return nil, nil, nil
}

var _ = Describe("AppServiceUnit", func() {
	Context("when the store has data", func() {
		var store *AppServiceStore
		var adapter *FakeAdapter
		var incomingChan chan appservice.AppServices
		var app1Service1 appservice.AppService

		BeforeEach(func() {
			adapter = &FakeAdapter{}
			c := cache.NewAppServiceCache()
			incomingChan = make(chan appservice.AppServices)
			app1Service1 = appservice.AppService{AppId: "app-1", Url: "syslog://example.com:12345"}
			store = NewAppServiceStore(adapter, c)

			go store.Run(incomingChan)
		})

		It("does not modify the store, when deleting data that doesn't exist", func() {
			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{app1Service1.Url},
			}

			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{},
			}

			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{},
			}
			Ω(adapter.DeleteCount).Should(Equal(1))
		})
	})
})
