package srvtopo

import (
	"context"
	"html/template"
	"sort"
	"time"

	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

// TopoTemplate is the HTML to use to display the
// ResilientServerCacheStatus object
const TopoTemplate = `
<style>
  table {
    border-collapse: collapse;
  }
  td, th {
    border: 1px solid #999;
    padding: 0.2rem;
  }
</style>
<table>
  <tr>
    <th colspan="4">SrvKeyspace Names Cache</th>
  </tr>
  <tr>
    <th>Cell</th>
    <th>SrvKeyspace Names</th>
    <th>TTL</th>
    <th>Error</th>
  </tr>
  {{range $i, $skn := .SrvKeyspaceNames}}
  <tr>
    <td>{{github_com_vitessio_vitess_vtctld_srv_cell $skn.Cell}}</td>
    <td>{{range $j, $value := $skn.Value}}{{github_com_vitessio_vitess_vtctld_srv_keyspace $skn.Cell $value}}&nbsp;{{end}}</td>
    <td>{{github_com_vitessio_vitess_srvtopo_ttl_time $skn.ExpirationTime}}</td>
    <td>{{if $skn.LastError}}({{github_com_vitessio_vitess_srvtopo_time_since $skn.LastQueryTime}}Ago) {{$skn.LastError}}{{end}}</td>
  </tr>
  {{end}}
</table>
<br>
<table>
  <tr>
    <th colspan="5">SrvKeyspace Cache</th>
  </tr>
  <tr>
    <th>Cell</th>
    <th>Keyspace</th>
    <th>SrvKeyspace</th>
    <th>TTL</th>
    <th>Error</th>
  </tr>
  {{range $i, $sk := .SrvKeyspaces}}
  <tr>
    <td>{{github_com_vitessio_vitess_vtctld_srv_cell $sk.Cell}}</td>
    <td>{{github_com_vitessio_vitess_vtctld_srv_keyspace $sk.Cell $sk.Keyspace}}</td>
    <td>{{$sk.StatusAsHTML}}</td>
    <td>{{github_com_vitessio_vitess_srvtopo_ttl_time $sk.ExpirationTime}}</td>
    <td>{{if $sk.LastError}}({{github_com_vitessio_vitess_srvtopo_time_since $sk.LastErrorTime}} Ago) {{$sk.LastError}}{{end}}</td>
  </tr>
  {{end}}
</table>
`

// The next few structures and methods are used to get a displayable
// version of the cache in a status page.

// SrvKeyspaceNamesCacheStatus is the current value for SrvKeyspaceNames
type SrvKeyspaceNamesCacheStatus struct {
	Cell           string
	Value          []string
	ExpirationTime time.Time
	LastQueryTime  time.Time
	LastError      error
	LastErrorCtx   context.Context
}

// SrvKeyspaceNamesCacheStatusList is used for sorting
type SrvKeyspaceNamesCacheStatusList []*SrvKeyspaceNamesCacheStatus

// Len is part of sort.Interface
func (skncsl SrvKeyspaceNamesCacheStatusList) Len() int {
	return len(skncsl)
}

// Less is part of sort.Interface
func (skncsl SrvKeyspaceNamesCacheStatusList) Less(i, j int) bool {
	return skncsl[i].Cell < skncsl[j].Cell
}

// Swap is part of sort.Interface
func (skncsl SrvKeyspaceNamesCacheStatusList) Swap(i, j int) {
	skncsl[i], skncsl[j] = skncsl[j], skncsl[i]
}

// SrvKeyspaceCacheStatus is the current value for a SrvKeyspace object
type SrvKeyspaceCacheStatus struct {
	Cell           string
	Keyspace       string
	Value          *topodatapb.SrvKeyspace
	ExpirationTime time.Time
	LastErrorTime  time.Time
	LastError      error
	LastErrorCtx   context.Context
}

// StatusAsHTML returns an HTML version of our status.
// It works best if there is data in the cache.
func (st *SrvKeyspaceCacheStatus) StatusAsHTML() template.HTML {
	if st.Value == nil {
		return template.HTML("No Data")
	}

	result := "<b>Partitions:</b><br>"
	for _, keyspacePartition := range st.Value.Partitions {
		result += "&nbsp;<b>" + keyspacePartition.ServedType.String() + ":</b>"
		for _, shard := range keyspacePartition.ShardReferences {
			result += "&nbsp;" + shard.Name
		}
		result += "<br>"
	}

	if st.Value.ShardingColumnName != "" {
		result += "<b>ShardingColumnName:</b>&nbsp;" + st.Value.ShardingColumnName + "<br>"
		result += "<b>ShardingColumnType:</b>&nbsp;" + st.Value.ShardingColumnType.String() + "<br>"
	}

	if len(st.Value.ServedFrom) > 0 {
		result += "<b>ServedFrom:</b><br>"
		for _, sf := range st.Value.ServedFrom {
			result += "&nbsp;<b>" + sf.TabletType.String() + ":</b>&nbsp;" + sf.Keyspace + "<br>"
		}
	}

	return template.HTML(result)
}

// SrvKeyspaceCacheStatusList is used for sorting
type SrvKeyspaceCacheStatusList []*SrvKeyspaceCacheStatus

// Len is part of sort.Interface
func (skcsl SrvKeyspaceCacheStatusList) Len() int {
	return len(skcsl)
}

// Less is part of sort.Interface
func (skcsl SrvKeyspaceCacheStatusList) Less(i, j int) bool {
	return skcsl[i].Cell+"."+skcsl[i].Keyspace <
		skcsl[j].Cell+"."+skcsl[j].Keyspace
}

// Swap is part of sort.Interface
func (skcsl SrvKeyspaceCacheStatusList) Swap(i, j int) {
	skcsl[i], skcsl[j] = skcsl[j], skcsl[i]
}

// ResilientServerCacheStatus has the full status of the cache
type ResilientServerCacheStatus struct {
	SrvKeyspaceNames SrvKeyspaceNamesCacheStatusList
	SrvKeyspaces     SrvKeyspaceCacheStatusList
}

// CacheStatus returns a displayable version of the cache
func (server *ResilientServer) CacheStatus() *ResilientServerCacheStatus {
	result := &ResilientServerCacheStatus{
		SrvKeyspaceNames: server.srvKeyspaceNamesQuery.CacheStatus(),
		SrvKeyspaces:     server.srvKeyspaceWatcher.CacheStatus(),
	}
	sort.Sort(result.SrvKeyspaceNames)
	sort.Sort(result.SrvKeyspaces)
	return result
}

// Returns the ttl for the cached entry or "Expired" if it is in the past
func ttlTime(expirationTime time.Time) template.HTML {
	ttl := time.Until(expirationTime).Round(time.Second)
	if ttl < 0 {
		return template.HTML("<b>Expired</b>")
	}
	return template.HTML(ttl.String())
}

func timeSince(t time.Time) template.HTML {
	return template.HTML(time.Since(t).Round(time.Second).String())
}

// StatusFuncs is required for CacheStatus) to work properly.
// We don't register them inside servenv directly so we don't introduce
// a dependency here.
var StatusFuncs = template.FuncMap{
	"github_com_vitessio_vitess_srvtopo_ttl_time":   ttlTime,
	"github_com_vitessio_vitess_srvtopo_time_since": timeSince,
}
