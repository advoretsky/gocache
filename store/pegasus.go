package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/XiaoMi/pegasus-go-client/admin"
	"github.com/XiaoMi/pegasus-go-client/pegasus"
	"github.com/spf13/cast"
	"strings"
	"time"
)

const (
	// PegasusType represents the storage type as a string value
	PegasusType = "pegasus"
	// PegasusTagPattern represents the tag pattern to be used as a key in specified storage
	PegasusTagPattern = "gocache_tag_%s"
	// Pegasus ttl(time-to-live) in seconds: -1 if ttl is not set; -2 if entry doesn't exist
	PegasusNOTTL   = -1
	PegasusNOENTRY = -2

	DefaultTable             = "gocache_pegasus"
	DefaultTablePartitionNum = 4
	DefaultScanNum           = 100
)

var (
	// empty represent empty sort key, more info reference: https://github.com/XiaoMi/pegasus-go-client/blob/f3b6b08bc4c227982bb5b73106329435fda97a38/pegasus/table_connector.go#L83
	empty = []byte("-")
)

// OptionsPegasus is options of Pegasus
type OptionsPegasus struct {
	Options
	MetaServers []string

	TableName         string
	TablePartitionNum int
	TableScanNum      int
}

// PegasusStore is a store for Pegasus
type PegasusStore struct {
	client  pegasus.Client
	options *OptionsPegasus
}

// NewPegasus creates a new store to pegasus instance(s)
func NewPegasus(options *OptionsPegasus) (*PegasusStore, error) {
	if options == nil {
		options = &OptionsPegasus{}
	}

	if err := createTable(options); err != nil {
		return nil, err
	}

	client := pegasus.NewClient(pegasus.Config{
		MetaServers: options.MetaServers,
	})
	table, err := client.OpenTable(options.Ctx, options.TableName)
	defer table.Close()
	if err != nil {
		return nil, err
	}

	return &PegasusStore{
		client:  client,
		options: options,
	}, nil
}

// validateOptions validate pegasus options
func validateOptions(options *OptionsPegasus) error {
	if len(options.MetaServers) == 0 {
		return errors.New("pegasus meta servers must fill")
	}
	if len(options.TableName) == 0 {
		options.TableName = DefaultTable
	}
	if options.TablePartitionNum < 1 {
		options.TablePartitionNum = DefaultTablePartitionNum
	}
	if options.TableScanNum < 1 {
		options.TableScanNum = DefaultScanNum
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	return nil
}

// createTable for create table by options
func createTable(options *OptionsPegasus) error {
	if err := validateOptions(options); err != nil {
		return err
	}

	tableClient := admin.NewClient(admin.Config{MetaServers: options.MetaServers})
	tableList, err := tableClient.ListTables(options.Ctx)
	if err != nil {
		return err
	}

	for i := range tableList {
		if tableList[i].Name == options.TableName {
			return nil
		}
	}

	// if not found then create table of options
	return tableClient.CreateTable(options.Ctx, options.TableName, options.TablePartitionNum)
}

// dropTable for drop table
func dropTable(options *OptionsPegasus) error {
	if err := validateOptions(options); err != nil {
		return err
	}

	tableClient := admin.NewClient(admin.Config{MetaServers: options.MetaServers})
	return tableClient.DropTable(options.Ctx, options.TableName)
}

// Close when exit store
func (p *PegasusStore) Close() error {
	return p.client.Close()
}

// Get returns data stored from a given key
func (p *PegasusStore) Get(key interface{}) (interface{}, error) {
	table, err := p.client.OpenTable(p.options.Ctx, p.options.TableName)
	defer table.Close()
	if err != nil {
		return nil, err
	}

	value, err := table.Get(p.options.Ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (p *PegasusStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	table, err := p.client.OpenTable(p.options.Ctx, p.options.TableName)
	defer table.Close()
	if err != nil {
		return nil, 0, err
	}

	value, err := table.Get(p.options.Ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, 0, err
	}

	ttl, err := table.TTL(p.options.Ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, 0, err
	}

	return value, time.Duration(ttl) * time.Second, nil
}

// Set defines data in Pegasus for given key identifier
func (p *PegasusStore) Set(key, value interface{}, options *Options) error {
	if options == nil {
		options = &Options{}
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	table, err := p.client.OpenTable(options.Ctx, p.options.TableName)
	defer table.Close()
	if err != nil {
		return err
	}

	err = table.SetTTL(options.Ctx, []byte(cast.ToString(key)), empty, []byte(cast.ToString(value)), options.Expiration)
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		if err = p.setTags(key, tags); err != nil {
			return err
		}
	}
	return nil
}

func (p *PegasusStore) setTags(key interface{}, tags []string) error {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(PegasusTagPattern, tag)
		var cacheKeys = []string{}

		if result, err := p.Get(tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		var alreadyInserted = false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		if err := p.Set(tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{
			Expiration: 720 * time.Hour,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes data from Pegasus for given key identifier
func (p *PegasusStore) Delete(key interface{}) error {
	table, err := p.client.OpenTable(p.options.Ctx, p.options.TableName)
	defer table.Close()
	if err != nil {
		return err
	}

	return table.Del(p.options.Ctx, []byte(cast.ToString(key)), empty)
}

// Invalidate invalidates some cache data in Pegasus for given options
func (p *PegasusStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(PegasusTagPattern, tag)
			result, err := p.Get(tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys = []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				if err := p.Delete(cacheKey); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (p *PegasusStore) Clear() error {
	table, err := p.client.OpenTable(p.options.Ctx, p.options.TableName)
	defer table.Close()
	if err != nil {
		return err
	}

	// init full scan
	scanners, err := table.GetUnorderedScanners(p.options.Ctx, p.options.TablePartitionNum, &pegasus.ScannerOptions{
		BatchSize: p.options.TableScanNum,
		// Values can be optimized out during scanning to reduce the workload.
		NoValue: true,
	})
	if err != nil {
		return err
	}

	// full scan and delete
	for _, scanner := range scanners {
		// Iterates sequentially.
		for true {
			completed, hashKey, _, _, err := scanner.Next(p.options.Ctx)
			if err != nil {
				return err
			}
			if completed {
				break
			}
			err = p.Delete(hashKey)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetType returns the store type
func (p *PegasusStore) GetType() string {
	return PegasusType
}
