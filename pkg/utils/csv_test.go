package utils

import (
	"github.com/google/go-cmp/cmp"
	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path"
	"reflect"
	"testing"

	"github.com/blang/semver"
	"github.com/operator-framework/operator-registry/pkg/registry"
)

func TestVerifyManifestDirs(t *testing.T) {
	type args struct {
		dirs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid manifest dir",
			args:    args{[]string{"./testdata/validManifests/3scale"}},
			wantErr: false,
		},
		{
			name:    "multiple valid manifest dirs",
			args:    args{[]string{"./testdata/validManifests/3scale", "./testdata/validManifests/3scale2"}},
			wantErr: false,
		},
		{
			name:    "invalid manifest dir no package.yaml",
			args:    args{[]string{"./testdata"}},
			wantErr: true,
		},
		{
			name:    "invalid manifest dir missing dir",
			args:    args{[]string{"./testdataaaaaaa"}},
			wantErr: true,
		},
		{
			name:    "multiple  invalid",
			args:    args{[]string{"./testdata", "./testdata/validManifests/3scale"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyManifestDirs(tt.args.dirs...); (err != nil) != tt.wantErr {
				t.Errorf("VerifyManifestDirs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPackageManifest(t *testing.T) {
	type args struct {
		packageDir string
	}
	tests := []struct {
		name    string
		args    args
		want    *registry.PackageManifest
		want1   string
		wantErr bool
	}{
		{
			name:    "valid package dir",
			args:    args{"./testdata/validManifests/3scale"},
			wantErr: false,
			want: &registry.PackageManifest{
				PackageName: "rhmi-3scale",
				Channels: []registry.PackageChannel{
					{
						Name:           "rhmi",
						CurrentCSVName: "3scale-operator.v0.4.0",
					},
				},
			},
			want1: "testdata/validManifests/3scale/3scale.package.yaml",
		},
		{
			name:    "valid package dir 2",
			args:    args{"./testdata/validManifests/3scale2"},
			wantErr: false,
			want: &registry.PackageManifest{
				PackageName: "rhmi-3scale",
				Channels: []registry.PackageChannel{
					{
						Name:           "rhmi",
						CurrentCSVName: "3scale-operator.v0.5.0",
					},
				},
			},
			want1: "testdata/validManifests/3scale2/3scale.package.yaml",
		},
		{
			name:    "invalid package dir",
			args:    args{"./testdata"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetPackageManifest(tt.args.packageDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackageManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPackageManifest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetPackageManifest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetSortedCSVNames(t *testing.T) {
	sortedcsvs := []csvName{
		{
			Name: "3scale-operator.v0.4.0",
			Version: semver.Version{
				Major: 0,
				Minor: 4,
				Patch: 0,
			},
		},
		{
			Name: "3scale-operator.v0.5.0",
			Version: semver.Version{
				Major: 0,
				Minor: 5,
				Patch: 0,
			},
		},
	}
	type args struct {
		packageDir string
	}
	tests := []struct {
		name    string
		args    args
		want    csvNames
		wantErr bool
	}{
		{
			name:    "valid get sorted dir",
			args:    args{"./testdata/validManifests/3scale2"},
			want:    sortedcsvs,
			wantErr: false,
		},
		{
			name:    "invalid package dir",
			args:    args{"./testdata/validManifests/somebaddir"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSortedCSVNames(tt.args.packageDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != nil {
					if got[0].Name != tt.want[0].Name {
						t.Errorf("GetCurrentCSV() got1 = %v, want %v", got, tt.want)
					}
				} else {
					t.Errorf("GetCurrentCSV() got = %v", got)
				}
			}
		})
	}

}

func TestGetCurrentCSV(t *testing.T) {
	type args struct {
		packageDir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "valid package dir",
			args:    args{"./testdata/validManifests/3scale"},
			wantErr: false,
			want:    "3scale-operator.v0.4.0",
			want1:   "testdata/validManifests/3scale/0.4.0/3scale-operator.v0.4.0.clusterserviceversion.yaml",
		},
		{
			name:    "valid package dir 2",
			args:    args{"./testdata/validManifests/3scale2"},
			wantErr: false,
			want:    "3scale-operator.v0.5.0",
			want1:   "testdata/validManifests/3scale2/0.5.0/3scale-operator.v0.5.0.clusterserviceversion.yaml",
		},
		{
			name:    "invalid package dir",
			args:    args{"./testdata"},
			wantErr: true,
		},
		{
			name:    "valid v2 bundle dir",
			args:    args{"./testdata/validManifests/v2/amq-online"},
			wantErr: false,
			want:    "amq-online.1.4.1",
			want1:   "testdata/validManifests/v2/amq-online/amq-online.1.4.1.clusterserviceversion.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetCurrentCSV(tt.args.packageDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != nil {
					if got.GetName() != tt.want {
						t.Errorf("GetCurrentCSV() got1 = %v, want %v", got.GetName(), tt.want)
					}
				} else {
					t.Errorf("GetCurrentCSV() got = %v", got)
				}
			}

			if got1 != tt.want1 {
				t.Errorf("GetCurrentCSV() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestReadCSVFromBundleDirectory(t *testing.T) {
	type args struct {
		bundleDir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "valid bundle dir",
			args:    args{"./testdata/validManifests/3scale/0.4.0"},
			wantErr: false,
			want:    "3scale-operator.v0.4.0",
			want1:   "testdata/validManifests/3scale/0.4.0/3scale-operator.v0.4.0.clusterserviceversion.yaml",
		},
		{
			name:    "invalid bundle dir",
			args:    args{"./testdata/validManifests/3scale"},
			wantErr: true,
		},
		{
			name:    "invalid dir",
			args:    args{"./testdataaaaaaaaaa"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ReadCSVFromBundleDirectory(tt.args.bundleDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCSVFromBundleDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != nil {
					if got.GetName() != tt.want {
						t.Errorf("ReadCSVFromBundleDirectory() got1 = %v, want %v", got.GetName(), tt.want)
					}
				} else {
					t.Errorf("ReadCSVFromBundleDirectory() got = %v", got)
				}
			}
			if got1 != tt.want1 {
				t.Errorf("ReadCSVFromBundleDirectory() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUnknownFieldsAreKept(t *testing.T) {
	src := "./testdata/csv/crw-2.1.0-csv.yaml"
	dest, err := ioutil.TempDir("/tmp", "csv-test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	csv, err := NewCSV(src)
	if err != nil {
		t.Fatalf("can not read csv file from %s. error: %v", src, err)
	}
	relatedImages, _, _ := unstructured.NestedSlice(csv.obj.Object, "spec", "relatedImages")
	if len(relatedImages) < 1 {
		t.Fatalf("there should be at least 1 image in relatedImages section")
	}
	out := path.Join(dest, "out.yaml")
	t.Logf("output file: %s", out)
	err = csv.WriteYAML(out)

	newCsv, err := NewCSV(out)
	if err != nil {
		t.Fatalf("can not read csv file from %s. error: %v", src, err)
	}
	newRelatedImages, _, err := unstructured.NestedSlice(newCsv.obj.Object, "spec", "relatedImages")
	if err != nil {
		t.Fatalf("can not find relatedImages field in the output file: %v", err)
	}
	if diff := cmp.Diff(relatedImages, newRelatedImages); diff != "" {
		t.Fatalf("unexpected diff: %s", diff)
	}
}

func TestCSVGetters(t *testing.T) {
	src := "./testdata/csv/crw-2.1.0-csv.yaml"
	csv, err := NewCSV(src)
	if err != nil {
		t.Fatalf("can not read csv file from %s. error: %v", src, err)
	}

	wantedVersion := "2.1.0"
	v, err := csv.GetVersion()
	if err != nil {
		t.Fatalf("failed to call GetVersion: %v", err)
	}
	if v.String() != wantedVersion {
		t.Fatalf("version value doesn't match. Wanted: %s, actual: %s", wantedVersion, v.String())
	}

	wantedName := "crwoperator.v2.1.0"
	if csv.GetName() != wantedName {
		t.Fatalf("name doesn't match. Wanted: %s, actual: %s", wantedName, csv.GetName())
	}

	wantedAnnotationsLength := 9
	if len(csv.GetAnnotations()) != wantedAnnotationsLength {
		t.Fatalf("annotations doesn't match. Wanted: %d, actual: %d", wantedAnnotationsLength, len(csv.GetAnnotations()))
	}

	expectedDeploymentSpecLengh := 1
	expectedDeploymentSpecName := "codeready-operator"
	deployments, err := csv.GetDeploymentSpecs()
	if err != nil {
		t.Fatalf("failed to call GetDeploymentSpecs: %v", err)
	}
	if len(deployments) != expectedDeploymentSpecLengh {
		t.Fatalf("deploymentspecs doesn't match. Wanted: %d, actual: %d", expectedDeploymentSpecLengh, len(deployments))
	}
	if deployments[0].Name != expectedDeploymentSpecName {
		t.Fatalf("deploymentspec name doesn't match. Wanted: %s, actual: %s", expectedDeploymentSpecName, deployments[0].Name)
	}

	d, err := csv.GetOperatorDeploymentSpec()
	if err != nil {
		t.Fatalf("failed to call GetOperatorDeploymentSpec: %v", err)
	}
	if d.Name != expectedDeploymentSpecName {
		t.Fatalf("deploymentspec name doesn't match. Wanted: %s, actual: %s", expectedDeploymentSpecName, d.Name)
	}
}

func TestCSVSetters(t *testing.T) {
	src := "./testdata/csv/crw-2.1.0-csv.yaml"
	csv, err := NewCSV(src)
	if err != nil {
		t.Fatalf("can not read csv file from %s. error: %v", src, err)
	}
	replacesValue := "test-2.0.0"
	if err := csv.SetReplaces(replacesValue); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	value, _, err := unstructured.NestedFieldNoCopy(csv.obj.Object, "spec", "replaces")
	if err != nil {
		t.Fatalf("can not get replaces field: %v", err)
	}
	if value != replacesValue {
		t.Fatalf("SetReplaces failed. Wanted: %s, actual: %s", replacesValue, value)
	}

	newAnnotations := map[string]string{
		"example": "test",
	}
	csv.SetAnnotations(newAnnotations)
	anno := csv.GetAnnotations()
	if diff := cmp.Diff(newAnnotations, anno); diff != "" {
		t.Fatalf("SetAnnotations failed. Diff: %s", diff)
	}

	newDeploymentSpec := olmapiv1alpha1.StrategyDeploymentSpec{
		Name: "test-spec",
		Spec: appv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-spec",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name: "namespace",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	existingDeploymentSpecs, err := csv.GetDeploymentSpecs()
	if err != nil {
		t.Fatalf("GetDeploymentSpecs failed. Error: %v", err)
	}
	deploymentSpecs := append(existingDeploymentSpecs, newDeploymentSpec)
	if err := csv.SetDeploymentSpecs(deploymentSpecs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotDeploymentSpecs, err := csv.GetDeploymentSpecs()
	if err != nil {
		t.Fatalf("GetDeploymentSpecs failed. Error: %v", err)
	}

	if diff := cmp.Diff(deploymentSpecs, gotDeploymentSpecs); diff != "" {
		t.Fatalf("DeploymentSpec doesn't match. Diff: %s", diff)
	}

	envVars := map[string]string{
		"namespace": "metadata.annotations['olm.targetNamespaces']",
	}

	if err := csv.UpdateEnvVarList(envVars); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	specs, _ := csv.GetDeploymentSpecs()
	updatedEnvVar := specs[1].Spec.Template.Spec.Containers[0].Env[0].ValueFrom.FieldRef.FieldPath
	if updatedEnvVar != envVars["namespace"] {
		t.Fatalf("UpdateEnvVarList failed. Wanted: %s, actual: %s", envVars["namespace"], updatedEnvVar)
	}

	if err := csv.SetOperatorDeploymentSpec(&newDeploymentSpec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, e := csv.GetOperatorDeploymentSpec()
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
	if diff := cmp.Diff(d, &newDeploymentSpec); diff != "" {
		t.Fatalf("SetOperatorDeploymentSpec failed. Diff = %s", diff)
	}
}
