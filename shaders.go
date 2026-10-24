package main

// Block shader. The layer of the tile comes in as the x of the second texture coordinate.
const (
	blockVertexShader = `#version 330
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec2 vertexTexCoord2;
in vec4 vertexColor;
uniform mat4 mvp;
out vec2 fragTexCoord;
flat out float fragLayer;
out vec4 fragColor;
void main() {
	fragTexCoord = vertexTexCoord;
	fragLayer = vertexTexCoord2.x;
	fragColor = vertexColor;
	gl_Position = mvp * vec4(vertexPosition, 1.0);
}
`
	blockFragmentShader = `#version 330
in vec2 fragTexCoord;
flat in float fragLayer;
in vec4 fragColor;
uniform sampler2DArray tiles;
out vec4 finalColor;
void main() {
	vec3 texel = texture(tiles, vec3(fragTexCoord, fragLayer)).rgb;
	finalColor = vec4(texel * fragColor.rgb, 1.0);
}
`
)
