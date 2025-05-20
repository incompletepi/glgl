#shader compute
#version 430

layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
layout(r32f, binding = 0) uniform image2D in_tex;
// The binding argument refers to the textures Unit.
layout(r32f, binding = 2) uniform image2D out_tex;

uniform float u_adder;

void main() {
    // get position to read/write data from.
    ivec2 pos = ivec2( gl_GlobalInvocationID.xy );
    // get value stored in the image
    float in_val = imageLoad( in_tex, pos ).r;
    // store new value in image
    imageStore( out_tex, pos, vec4( in_val + u_adder, 0.0, 0.0, 0.0 ) );
}