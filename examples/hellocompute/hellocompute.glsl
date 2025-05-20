#shader compute
#version 430

// These are some of the openGL defined inputs for compute shaders.
// in uvec3 gl_NumWorkGroups;
// in uvec3 gl_WorkGroupID;
// in uvec3 gl_LocalInvocationID;
// in uvec3 gl_GlobalInvocationID;
// in uint gl_LocalInvocationIndex;

layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
layout(r32f, binding = 0) uniform image2D out_tex;

uniform float u_adder;

void main() {
    // get position to read/write data from.
    ivec2 pos = ivec2( gl_GlobalInvocationID.xy );
    // get value stored in the image
    float in_val = imageLoad( out_tex, pos ).r;
    // store new value in image
    imageStore( out_tex, pos, vec4( in_val + u_adder, 0.0, 0.0, 0.0 ) );
}