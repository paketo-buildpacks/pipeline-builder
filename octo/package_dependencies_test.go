package octo

import (
	"github.com/buildpacks/libcnb"
	_package "github.com/paketo-buildpacks/pipeline-builder/octo/package"
	"testing"
)

func TestFindIds(t *testing.T) {
	type args struct {
		bpOrders   _package.BuildpackOrderGroups
		dep        _package.Dependency
		descriptor Descriptor
	}

	var tests = []struct {
		name        string
		args        args
		packageId   string
		buildpackId string
		wantErr     bool
	}{
		{
			name: "GroupId ends with buildpackId",
			args: struct {
				bpOrders   _package.BuildpackOrderGroups
				dep        _package.Dependency
				descriptor Descriptor
			}{
				bpOrders: _package.BuildpackOrderGroups{
					Orders: []libcnb.BuildpackOrder{
						{
							Groups: []libcnb.BuildpackOrderBuildpack{
								{
									ID:       "paketo-buildpacks/bellsoft-liberica",
									Version:  "9.12.0",
									Optional: false,
								},
							},
						},
					},
				},
				dep: struct {
					URI string
				}{
					URI: "docker://gcr.io/tanzu-buildpacks/bellsoft-liberica:9.12.0",
				},
				descriptor: Descriptor{},
			},
			packageId:   "gcr.io/tanzu-buildpacks/bellsoft-liberica",
			buildpackId: "gcr.io/paketo-buildpacks/bellsoft-liberica",
			wantErr:     false,
		},
		{
			name: "PackageId ends with -offline",
			args: struct {
				bpOrders   _package.BuildpackOrderGroups
				dep        _package.Dependency
				descriptor Descriptor
			}{
				bpOrders: _package.BuildpackOrderGroups{
					Orders: []libcnb.BuildpackOrder{
						{
							Groups: []libcnb.BuildpackOrderBuildpack{
								{
									ID:       "tanzu-buildpacks/tanzu-bellsoft-liberica",
									Version:  "9.12.0",
									Optional: false,
								},
							},
						},
					},
				},
				dep: struct {
					URI string
				}{
					URI: "docker://gcr.io/tanzu-buildpacks/tanzu-bellsoft-liberica-offline:9.12.0",
				},
				descriptor: Descriptor{},
			},
			packageId:   "gcr.io/tanzu-buildpacks/tanzu-bellsoft-liberica-offline",
			buildpackId: "gcr.io/tanzu-buildpacks/tanzu-bellsoft-liberica",
			wantErr:     false,
		},
		{
			name: "Unable to match coordinates",
			args: struct {
				bpOrders   _package.BuildpackOrderGroups
				dep        _package.Dependency
				descriptor Descriptor
			}{
				bpOrders: _package.BuildpackOrderGroups{
					Orders: []libcnb.BuildpackOrder{
						{
							Groups: []libcnb.BuildpackOrderBuildpack{
								{
									ID:       "tanzu-buildpacks/tanzu-bellsoft-liberica",
									Version:  "9.12.0",
									Optional: false,
								},
							},
						},
					},
				},
				dep: struct {
					URI string
				}{
					URI: "docker://gcr.io/tanzu-buildpacks/tanzu-bellsoft-liberica-HELLO:9.12.0",
				},
				descriptor: Descriptor{
					PackageMatcher: "-offline",
				},
			},
			packageId:   "",
			buildpackId: "",
			wantErr:     true,
		},
		{
			name: "PackageId is a perfect match",
			args: struct {
				bpOrders   _package.BuildpackOrderGroups
				dep        _package.Dependency
				descriptor Descriptor
			}{
				bpOrders: _package.BuildpackOrderGroups{
					Orders: []libcnb.BuildpackOrder{
						{
							Groups: []libcnb.BuildpackOrderBuildpack{
								{
									ID:       "tanzu-buildpacks/deprecation-warnings",
									Version:  "0.0.3",
									Optional: true,
								},
							},
						},
					},
				},
				dep: struct {
					URI string
				}{
					URI: "docker://gcr.io/tanzu-buildpacks/deprecation-warnings:0.0.3",
				},
				descriptor: Descriptor{},
			},
			packageId:   "gcr.io/tanzu-buildpacks/deprecation-warnings",
			buildpackId: "gcr.io/tanzu-buildpacks/deprecation-warnings",
			wantErr:     false,
		},
		{
			name: "Dep URI is not valid",
			args: struct {
				bpOrders   _package.BuildpackOrderGroups
				dep        _package.Dependency
				descriptor Descriptor
			}{
				bpOrders: _package.BuildpackOrderGroups{
					Orders: []libcnb.BuildpackOrder{
						{
							Groups: []libcnb.BuildpackOrderBuildpack{
								{
									ID:       "tanzu-buildpacks/deprecation-warnings",
									Version:  "0.0.3",
									Optional: true,
								},
							},
						},
					},
				},
				dep: struct {
					URI string
				}{
					URI: "hello all!",
				},
				descriptor: Descriptor{},
			},
			packageId:   "",
			buildpackId: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := findIds(tt.args.bpOrders, tt.args.dep, tt.args.descriptor)
			if (err != nil) != tt.wantErr {
				t.Errorf("findIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.packageId {
				t.Errorf("findIds() got = %v, packageId %v", got, tt.packageId)
			}
			if got1 != tt.buildpackId {
				t.Errorf("findIds() got1 = %v, packageId %v", got1, tt.buildpackId)
			}
		})
	}
}
