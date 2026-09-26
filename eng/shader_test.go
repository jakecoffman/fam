package eng

import "testing"

func TestShaderCachedUniformLocations(t *testing.T) {
	shader := &Shader{uniforms: map[string]int32{"first": 0, "other": 3, "missing": -1}}
	for name, want := range shader.uniforms {
		if got := shader.uniformLocation(name); got != want {
			t.Errorf("uniform %q = %d, want cached %d", name, got, want)
		}
	}
}
