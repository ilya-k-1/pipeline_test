package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type FunctionDescription struct {
	Defs        map[string]string   `json:"definitions"`
	Inputs      map[string][]string `json:"inputs"`
	Descendants map[string][]string `json:"descendants"`
	Arguments   []string            `json:"arguments"`
}

func map_names_vn(lst []ssa.Value, dict map[ssa.Value]string) []string {
	res := []string{}
	for index := range lst {
		res = append(res, dict[lst[index]])
	}
	return res
}

func map_names_pcode(lst []ssa.Instruction, dict map[ssa.Instruction]string) []string {
	res := []string{}
	for index := range lst {
		res = append(res, dict[lst[index]])
	}
	return res
}

func serialize_function(m ssa.Member) (FunctionDescription, string) {
	defs := make(map[string]string)
	inputs := make(map[string][]string)
	descendants := make(map[string][]string)
	arguments := make([]string, 0)
	fname := ""

	if m.Token() == token.FUNC {
		f := m.(*ssa.Function)
		fname = f.Name()
		descendants_map := make(map[ssa.Value][]ssa.Instruction)
		def_map := make(map[ssa.Value]ssa.Instruction)
		inputs_map := make(map[ssa.Instruction][]ssa.Value)

		varnodes := make(map[ssa.Value]string)
		pcodes := make(map[ssa.Instruction]string)

		varnodes_counter := 0
		pcodes_counter := 0

		for arg_index := range f.Params {
			arguments = append(arguments, f.Params[arg_index].Name())
		}

		for block := range f.Blocks {
			for instr := range f.Blocks[block].Instrs {
				instruction := f.Blocks[block].Instrs[instr]
				pcodes[instruction] = instruction.String() + fmt.Sprintf("##%d", pcodes_counter)
				pcodes_counter++
				instr_val, ok := instruction.(ssa.Value)
				if ok {
					varnodes[instr_val] = instr_val.String() + fmt.Sprintf("##%d", varnodes_counter)
					varnodes_counter++
					def_map[instr_val] = instruction
				}

				var op_slice []*ssa.Value
				op_slice = f.Blocks[block].Instrs[instr].Operands(op_slice)
				for op := range op_slice {
					vn := *op_slice[op]
					if vn != nil {
						_, ok = varnodes[vn]
						if !ok {
							varnodes[vn] = vn.String() + fmt.Sprintf("##%d", varnodes_counter)
							varnodes_counter++
						}
						descendants_map[vn] = append(descendants_map[vn], instruction)
						inputs_map[instruction] = append(inputs_map[instruction], vn)
					}
				}
			}
		}

		for k, v := range def_map {
			defs[varnodes[k]] = pcodes[v]
		}

		for k, v := range inputs_map {
			inputs[pcodes[k]] = map_names_vn(v, varnodes)
		}

		for k, v := range descendants_map {
			descendants[varnodes[k]] = map_names_pcode(v, pcodes)
		}

	}
	return FunctionDescription{defs, inputs, descendants, arguments}, fname

}

func main() {
	// Read the stdin

	// Parse the source files.

	var fname_in string
	var fname_out string
	flag.StringVar(&fname_in, "in", "", "Input golang file")
	flag.StringVar(&fname_out, "out", "", "Output golang file")
	flag.Parse()
	if fname_in == "" {
		fmt.Println("File name missing")
		return
	}
	abspath, _ := filepath.Abs(fname_in)
	dat, err := ioutil.ReadFile(abspath)

	if err == nil {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "pack.go", dat, parser.ParseComments)
		if err != nil {
			fmt.Print(err) // parse error
			return
		}
		files := []*ast.File{f}

		// Create the type-checker's package.
		pkg := types.NewPackage("pack", "")

		// Type-check the package, load dependencies.
		// Create and build the SSA program.
		digest, _, err := ssautil.BuildPackage(
			&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
		if err != nil {
			fmt.Print(err) // type error in some package
			return
		}
		serialization_data := make(map[string]FunctionDescription)
		for fun := range digest.Members {
			ser, name := serialize_function(digest.Members[fun])
			serialization_data[name] = ser
		}
		out, err := json.Marshal(serialization_data)
		if err == nil {
			ioutil.WriteFile(fname_out, out, 0644)
		} else {
			fmt.Println("Serialization failed")
		}
	} else {
		fmt.Println("Could not read file")
	}

}
