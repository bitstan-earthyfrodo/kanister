// Copyright 2022 The Kanister Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"encoding/json"
	"reflect"

	"github.com/kopia/kopia/fs"
	"github.com/kopia/kopia/snapshot"
	. "gopkg.in/check.v1"
)

type KopiaParseUtilsTestSuite struct{}

var _ = Suite(&KopiaParseUtilsTestSuite{})

func (kParse *KopiaParseUtilsTestSuite) TestSnapshotIDsFromSnapshot(c *C) {
	for _, tc := range []struct {
		log            string
		expectedSnapID string
		expectedRootID string
		errChecker     Checker
	}{
		{"Created snapshot with root k23cf6d7ff418a0110636399da458abb5 and ID beda41fb4ba7478025778fdc8312355c in 10.8362ms", "beda41fb4ba7478025778fdc8312355c", "k23cf6d7ff418a0110636399da458abb5", IsNil},
		{"Created snapshot with root rootID and ID snapID", "snapID", "rootID", IsNil},
		{" Created snapshot snapID (root rootID)", "", "", NotNil},
		{"root 123abcd", "", "", NotNil},
		{"Invalid message", "", "", NotNil},
		{"Created snapshot with root abc123\n in 5.5001ms", "", "", NotNil},
		{"", "", "", NotNil},
		{"Created snapshot", "", "", NotNil},
		{"Created snapshot ", "", "", NotNil},
		{"Created snapshot with root", "", "", NotNil},
		{"Created snapshot with root rootID", "", "", NotNil},
		{"Created snapshot with root rootID and ID\n snapID in 10ms", "", "", NotNil},
		{"Created snapshot with root rootID in 10ms", "", "", NotNil},
		{"Created snapshot and ID snapID in 10ms", "", "", NotNil},
		{"Created snapshot with ID snapID in 10ms", "", "", NotNil},
		{"Created snapshot snapID\n(root rootID) in 10.8362ms", "", "", NotNil},
		{"Created snapshot snapID in 10.8362ms", "", "", NotNil},
		{"Created snapshot (root rootID) in 10.8362ms", "", "", NotNil},
		{"Created snapshot root rootID in 10.8362ms", "", "", NotNil},
		{"Created snapshot root rootID and ID snapID in 10.8362ms", "", "", NotNil},
		{" root rootID and ID snapID in 10.8362ms", "", "", NotNil},
		{"uploaded snapshot beda41fb4ba7478025778fdc8312355c (root k23cf6d7ff418a0110636399da458abb5) in 10.8362ms", "", "", NotNil},
	} {
		snapID, rootID, err := SnapshotIDsFromSnapshot(tc.log)
		c.Check(snapID, Equals, tc.expectedSnapID, Commentf("Failed for log: %s", tc.log))
		c.Check(rootID, Equals, tc.expectedRootID, Commentf("Failed for log: %s", tc.log))
		c.Check(err, tc.errChecker, Commentf("Failed for log: %s", tc.log))
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestLatestSnapshotInfoFromManifestList(c *C) {
	for _, tc := range []struct {
		output             string
		checker            Checker
		expectedSnapID     string
		expectedBackupPath string
	}{
		{
			output: `[
				{"id":"00000000000000000000001","length":604,"labels":{"hostname":"h2","path":"/tmp/aaa1","type":"snapshot","username":"u2"},"mtime":"2021-05-19T11:53:50.882509009Z"},
				{"id":"00000000000000000000002","length":603,"labels":{"hostname":"h2","path":"/tmp/aaa2","type":"snapshot","username":"u2"},"mtime":"2021-05-19T12:24:11.258017051Z"},
				{"id":"00000000000000000000003","length":602,"labels":{"hostname":"h2","path":"/tmp/aaa3","type":"snapshot","username":"u2"},"mtime":"2021-05-19T12:24:25.767315039Z"}
			   ]`,
			expectedSnapID:     "00000000000000000000003",
			expectedBackupPath: "/tmp/aaa3",
			checker:            IsNil,
		},
		{
			output:             "",
			expectedSnapID:     "",
			expectedBackupPath: "",
			checker:            NotNil,
		},
		{
			output: `[
				{"id":"","length":602,"labels":{"hostname":"h2","path":"/tmp/aaa3","type":"snapshot","username":"u2"},"mtime":"2021-05-19T12:24:25.767315039Z"}
			   ]`,
			expectedSnapID:     "",
			expectedBackupPath: "",
			checker:            NotNil,
		},
		{
			output: `[
				{"id":"00000000000000000000003","length":602,"labels":{"hostname":"h2","path":"","type":"snapshot","username":"u2"},"mtime":"2021-05-19T12:24:25.767315039Z"}
			   ]`,
			expectedSnapID:     "",
			expectedBackupPath: "",
			checker:            NotNil,
		},
	} {
		snapID, backupPath, err := LatestSnapshotInfoFromManifestList(tc.output)
		c.Assert(err, tc.checker)
		c.Assert(snapID, Equals, tc.expectedSnapID)
		c.Assert(backupPath, Equals, tc.expectedBackupPath)
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestSnapshotInfoFromSnapshotCreateOutput(c *C) {
	for _, tc := range []struct {
		output         string
		checker        Checker
		expectedSnapID string
		expectedRootID string
	}{
		{
			output: `Snapshotting u2@h2:/tmp/aaa1 ...
			* 0 hashing, 1 hashed (2 B), 3 cached (4 B), uploaded 5 KB, estimating...
		   {"id":"00000000000000000000001","source":{"host":"h2","userName":"u2","path":"/tmp/aaa1"},"description":"","startTime":"2021-05-26T05:29:07.206854927Z","endTime":"2021-05-26T05:29:07.207328392Z","rootEntry":{"name":"aaa1","type":"d","mode":"0755","mtime":"2021-05-19T15:45:34.448853232Z","obj":"ka68ba7abe0818b24a2b0647aeeb02f29","summ":{"size":0,"files":1,"symlinks":0,"dirs":1,"maxTime":"2021-05-19T15:45:34.448853232Z","numFailed":0}}}
		   `,
			checker:        IsNil,
			expectedSnapID: "00000000000000000000001",
			expectedRootID: "ka68ba7abe0818b24a2b0647aeeb02f29",
		},
		{
			output: `Snapshotting u2@h2:/tmp/aaa1 ...
			* 0 hashing, 1 hashed (2 B), 3 cached (4 B), uploaded 5 KB, estimating...
		   `,
			checker:        NotNil,
			expectedSnapID: "",
			expectedRootID: "",
		},
		{
			output: `ERROR: unable to get local filesystem entry: resolveSymlink: stat: lstat /tmp/aaa2: no such file or directory
			`,
			checker:        NotNil,
			expectedSnapID: "",
			expectedRootID: "",
		},
	} {
		snapID, rootID, err := SnapshotInfoFromSnapshotCreateOutput(tc.output)
		c.Assert(err, tc.checker)
		c.Assert(snapID, Equals, tc.expectedSnapID)
		c.Assert(rootID, Equals, tc.expectedRootID)
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestSnapSizeStatsFromSnapListAll(c *C) {
	for _, tc := range []struct {
		description     string
		outputGenFunc   func(*C, []*snapshot.Manifest) string
		expManifestList []*snapshot.Manifest
		expCount        int
		expSize         int64
		errChecker      Checker
	}{
		{
			description:     "empty manifest list",
			outputGenFunc:   marshalManifestList,
			expManifestList: []*snapshot.Manifest{},
			expCount:        0,
			expSize:         0,
			errChecker:      IsNil,
		},
		{
			description:   "basic manifest list",
			outputGenFunc: marshalManifestList,
			expManifestList: []*snapshot.Manifest{
				{
					RootEntry: &snapshot.DirEntry{
						DirSummary: &fs.DirectorySummary{
							TotalFileSize: 1,
						},
					},
				},
			},
			expCount:   1,
			expSize:    1,
			errChecker: IsNil,
		},
		{
			description:   "manifest list with multiple snapshots",
			outputGenFunc: marshalManifestList,
			expManifestList: []*snapshot.Manifest{
				{
					RootEntry: &snapshot.DirEntry{
						DirSummary: &fs.DirectorySummary{
							TotalFileSize: 1,
						},
					},
				},
				{
					RootEntry: &snapshot.DirEntry{
						DirSummary: &fs.DirectorySummary{
							TotalFileSize: 10,
						},
					},
				},
				{
					RootEntry: &snapshot.DirEntry{
						DirSummary: &fs.DirectorySummary{
							TotalFileSize: 100,
						},
					},
				},
				{
					RootEntry: &snapshot.DirEntry{
						DirSummary: &fs.DirectorySummary{
							TotalFileSize: 1000,
						},
					},
				},
			},
			expCount:   4,
			expSize:    1111,
			errChecker: IsNil,
		},
		{
			description:   "error: snapshot with no directory summary, size is treated as zero",
			outputGenFunc: marshalManifestList,
			expManifestList: []*snapshot.Manifest{
				{
					RootEntry: &snapshot.DirEntry{},
				},
			},
			expCount:   1,
			expSize:    0,
			errChecker: IsNil,
		},
		{
			description:   "error: snapshot with no root entry, size is treated as zero",
			outputGenFunc: marshalManifestList,
			expManifestList: []*snapshot.Manifest{
				{},
			},
			expCount:   1,
			expSize:    0,
			errChecker: IsNil,
		},
		{
			description: "error: parse empty output",
			outputGenFunc: func(c *C, manifestList []*snapshot.Manifest) string {
				return ""
			},
			expCount:   0,
			expSize:    0,
			errChecker: NotNil,
		},
		{
			description: "error: unmarshal fails",
			outputGenFunc: func(c *C, manifestList []*snapshot.Manifest) string {
				return "asdf"
			},
			expCount:   0,
			expSize:    0,
			errChecker: NotNil,
		},
	} {
		outputToParse := tc.outputGenFunc(c, tc.expManifestList)
		gotTotSizeB, gotNumSnapshots, err := SnapSizeStatsFromSnapListAll(outputToParse)
		c.Check(err, tc.errChecker, Commentf("Failed for output: %q", outputToParse))
		c.Check(gotTotSizeB, Equals, tc.expSize)
		c.Check(gotNumSnapshots, Equals, tc.expCount)
		c.Log(err)
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestSnapshotStatsFromSnapshotCreate(c *C) {
	type args struct {
		snapCreateOutput  string
		matchOnlyFinished bool
	}
	tests := []struct {
		name      string
		args      args
		wantStats *SnapshotCreateStats
	}{
		{
			name: "Basic test case",
			args: args{
				snapCreateOutput: " \\ 0 hashing, 1 hashed (2 B), 3 cached (40 KB), uploaded 6.7 GB, estimated 2044.2 MB (95.5%) 0s left",
			},
			wantStats: &SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     40000,
				SizeUploadedB:   6700000000,
				SizeEstimatedB:  2044200000,
				ProgressPercent: 95,
			},
		},
		{
			name: "Real test case",
			args: args{
				snapCreateOutput: " - 0 hashing, 283 hashed (219.5 MB), 0 cached (0 B), uploaded 10.5 MB, estimated 6.01 MB (91.7%) 1s left",
			},
			wantStats: &SnapshotCreateStats{
				FilesHashed:     283,
				SizeHashedB:     219500000,
				FilesCached:     0,
				SizeCachedB:     0,
				SizeUploadedB:   10500000,
				SizeEstimatedB:  6010000,
				ProgressPercent: 91,
			},
		},
		{
			name: "Check multiple digits each field",
			args: args{
				snapCreateOutput: " * 0 hashing, 123 hashed (1234.5 MB), 123 cached (1234 B), uploaded 1234.5 KB, estimated 941.2 KB (100.0%) 0s left",
			},
			wantStats: &SnapshotCreateStats{
				FilesHashed:     123,
				SizeHashedB:     1234500000,
				FilesCached:     123,
				SizeCachedB:     1234,
				SizeUploadedB:   1234500,
				SizeEstimatedB:  941200,
				ProgressPercent: 100,
			},
		},
		{
			name: "Ignore running output when expecting completed line",
			args: args{
				snapCreateOutput:  "| 0 hashing, 1 hashed (2 B), 3 cached (4 B), uploaded 5 KB, estimating...",
				matchOnlyFinished: true,
			},
			wantStats: nil,
		},
		{
			name: "Check estimating when running",
			args: args{
				snapCreateOutput: "| 0 hashing, 1 hashed (2 B), 3 cached (4 B), uploaded 5 KB, estimating...",
			},
			wantStats: &SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5000,
				SizeEstimatedB:  0,
				ProgressPercent: 0,
			},
		},
		{
			name: "Check estimating when finished",
			args: args{
				snapCreateOutput:  "* 0 hashing, 1 hashed (2 B), 3 cached (4 B), uploaded 5 KB, estimating...",
				matchOnlyFinished: true,
			},
			wantStats: &SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5000,
				SizeEstimatedB:  0,
				ProgressPercent: 100,
			},
		},
	}
	for _, tt := range tests {
		if gotStats := SnapshotStatsFromSnapshotCreate(tt.args.snapCreateOutput, tt.args.matchOnlyFinished); !reflect.DeepEqual(gotStats, tt.wantStats) {
			c.Errorf("SnapshotStatsFromSnapshotCreate() = %v, want %v", gotStats, tt.wantStats)
		}
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestPhysicalSizeFromBlobStatsRaw(c *C) {
	for _, tc := range []struct {
		blobStatsOutput string
		expSizeVal      int64
		expCount        int
		errChecker      Checker
	}{
		{
			"Count: 813\nTotal: 11235\n",
			11235,
			813,
			IsNil,
		},
		{
			"Total: 11235\nCount: 813\n",
			11235,
			813,
			IsNil,
		},
		{
			"Count: 0\nTotal: 0\n",
			0,
			0,
			IsNil,
		},
		{
			"Count: 5\nTotal: 0.0\n",
			0,
			0,
			NotNil,
		},
		{
			"Count: 5\nTotal: asdf\n",
			0,
			0,
			NotNil,
		},
		{
			"Count: 5\nTotal: 11235,\n",
			0,
			0,
			NotNil,
		},
		{
			"Total: -11235\n",
			0,
			0,
			NotNil,
		},
		{
			"Total: 11235",
			0,
			0,
			NotNil,
		},
		{
			"Count: 11235",
			0,
			0,
			NotNil,
		},
		{
			"Other-field: 11235",
			0,
			0,
			NotNil,
		},
		{
			"random input that doesn't comply with expected format",
			0,
			0,
			NotNil,
		},
		{
			`
Count: 26
Total: 65628
Average: 2524
Histogram:

		0 between 0 and 10 (total 0)
		0 between 10 and 100 (total 0)
		11 between 100 and 1000 (total 2132)
		15 between 1000 and 10000 (total 63496)
		0 between 10000 and 100000 (total 0)
		0 between 100000 and 1000000 (total 0)
		0 between 1000000 and 10000000 (total 0)
		0 between 10000000 and 100000000 (total 0)`,
			65628,
			26,
			IsNil,
		},
	} {
		gotSize, gotCount, err := RepoSizeStatsFromBlobStatsRaw(tc.blobStatsOutput)
		c.Check(err, tc.errChecker, Commentf("Failed for log: %s", tc.blobStatsOutput))
		c.Check(gotSize, Equals, tc.expSizeVal)
		c.Check(gotCount, Equals, tc.expCount)
	}
}

func (kParse *KopiaParseUtilsTestSuite) TestIsEqualSnapshotCreateStats(c *C) {
	for _, tc := range []struct {
		description string
		a           *SnapshotCreateStats
		b           *SnapshotCreateStats
		expResult   bool
	}{
		{
			"Both nil",
			nil,
			nil,
			true,
		},
		{
			"First nil",
			nil,
			&SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5,
				SizeEstimatedB:  6,
				ProgressPercent: 7,
			},
			false,
		},
		{
			"Second nil",
			&SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5,
				SizeEstimatedB:  6,
				ProgressPercent: 7,
			},
			nil,
			false,
		},
		{
			"Not nil, equal",
			&SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5,
				SizeEstimatedB:  6,
				ProgressPercent: 7,
			},
			&SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5,
				SizeEstimatedB:  6,
				ProgressPercent: 7,
			},
			true,
		},
		{
			"Not nil, not equal",
			&SnapshotCreateStats{
				FilesHashed:     1,
				SizeHashedB:     2,
				FilesCached:     3,
				SizeCachedB:     4,
				SizeUploadedB:   5,
				SizeEstimatedB:  6,
				ProgressPercent: 7,
			},
			&SnapshotCreateStats{
				FilesHashed:     5,
				SizeHashedB:     7,
				FilesCached:     2,
				SizeCachedB:     8,
				SizeUploadedB:   5,
				SizeEstimatedB:  2,
				ProgressPercent: 1,
			},
			false,
		},
	} {
		result := IsEqualSnapshotCreateStats(tc.a, tc.b)
		c.Check(result, Equals, tc.expResult)
	}
}

func marshalManifestList(c *C, manifestList []*snapshot.Manifest) string {
	c.Assert(manifestList, NotNil)

	b, err := json.Marshal(manifestList)
	c.Assert(err, IsNil)

	return string(b)
}
