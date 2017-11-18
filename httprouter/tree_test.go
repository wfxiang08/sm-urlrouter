// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//
// go test url_manager/httprouter -v -run "TestUrlTree$"
//
func TestUrlTree(t *testing.T) {

	urls := []string{
		`PUT api/v17/<platform>/<client>/<language>/<deviceType>/<deviceDensity>/users/<user_id:\d+>/images###user/image`,
		`POST api/v17/<platform>/<client>/<language>/<deviceType>/<deviceDensity>/users/<user_id:\d+>/songs/<song_id:\d+>###user/song-redeemed`,
	}
	root := &TreeNode{}

	for _, url := range urls {
		action, err := PreprocessPhpUrl(url)
		assert.NoError(t, err, "url parse error")
		fileds := SplitUrlSegments(action.Url)
		err = root.AddRoute(fileds, action)
		assert.NoError(t, err, "url add error")
	}
}
