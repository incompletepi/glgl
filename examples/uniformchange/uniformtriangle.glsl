// Separate our vertex from our fragment shaders
// with a #shader pragma. Taken from The Cherno's youtube OpenGL series.
#shader vertex
#version 330

in vec3 vert;

void main() {
	gl_Position = vec4(vert.xyz, 1.0);
}

#shader fragment
#version 330

out vec4 outputColor;

uniform vec4 u_color;

void main() {
	outputColor = u_color;
}