package tlc

import (
	"bytes"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type Resolver interface {
	TryResolveTypeName(name string, context string) Repr
	ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr
}

type Repr interface {
	Resolve(resolver Resolver)
	AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string)
	AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string)
	AppendSwitchCase(buf *bytes.Buffer, indent, reader string)
	AppendGoDefs(buf *bytes.Buffer)
	GoType() string
	GoImports() []string
}

type UnknownTypeRefRepr struct {
	Name string
}

func (r *UnknownTypeRefRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: read ")
	buf.WriteString(dst)
	buf.WriteString("\n")
}

func (r *UnknownTypeRefRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: write ")
	buf.WriteString(src)
	buf.WriteString("\n")
}

func (r *UnknownTypeRefRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *UnknownTypeRefRepr) Resolve(resolver Resolver) {
}

func (r *UnknownTypeRefRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *UnknownTypeRefRepr) GoType() string {
	return "interface{} /* " + r.Name + " */"
}
func (r *UnknownTypeRefRepr) GoImports() []string {
	return nil
}

type UndefinedRepr struct {
}

func (r *UndefinedRepr) Resolve(resolver Resolver) {
}

func (r *UndefinedRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: read ")
	buf.WriteString(dst)
	buf.WriteString("\n")
}

func (r *UndefinedRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: write ")
	buf.WriteString(src)
	buf.WriteString("\n")
}

func (r *UndefinedRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *UndefinedRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *UndefinedRepr) GoType() string {
	return "interface{}"
}
func (r *UndefinedRepr) GoImports() []string {
	return nil
}

type StringRepr struct {
}

func (r *StringRepr) Resolve(resolver Resolver) {
}

func (r *StringRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString(reader)
	buf.WriteString(".ReadString()")
	buf.WriteString("\n")
}

func (r *StringRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteString(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *StringRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *StringRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *StringRepr) GoType() string {
	return "string"
}
func (r *StringRepr) GoImports() []string {
	return nil
}

type BytesRepr struct {
}

func (r *BytesRepr) Resolve(resolver Resolver) {
}

func (r *BytesRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString(reader)
	buf.WriteString(".ReadBlog()")
	buf.WriteString("\n")
}

func (r *BytesRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteBlob(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *BytesRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *BytesRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *BytesRepr) GoType() string {
	return "[]byte"
}
func (r *BytesRepr) GoImports() []string {
	return nil
}

type BigIntRepr struct {
}

func (r *BigIntRepr) Resolve(resolver Resolver) {
}

func (r *BigIntRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString(reader)
	buf.WriteString(".ReadBigInt()")
	buf.WriteString("\n")
}

func (r *BigIntRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteBigInt(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *BigIntRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *BigIntRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *BigIntRepr) GoType() string {
	return "*big.Int"
}
func (r *BigIntRepr) GoImports() []string {
	return []string{"math/big"}
}

type IntRepr struct {
}

func (r *IntRepr) Resolve(resolver Resolver) {
}

func (r *IntRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString(reader)
	buf.WriteString(".ReadInt()")
	buf.WriteString("\n")
}

func (r *IntRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteInt(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *IntRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *IntRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *IntRepr) GoType() string {
	return "int"
}
func (r *IntRepr) GoImports() []string {
	return nil
}

type LongRepr struct {
}

func (r *LongRepr) Resolve(resolver Resolver) {
}

func (r *LongRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString(reader)
	buf.WriteString(".ReadUint64()")
	buf.WriteString("\n")
}

func (r *LongRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteUint64(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *LongRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *LongRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *LongRepr) GoType() string {
	return "uint64"
}
func (r *LongRepr) GoImports() []string {
	return nil
}

type Int128Repr struct {
}

func (r *Int128Repr) Resolve(resolver Resolver) {
}

func (r *Int128Repr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(reader)
	buf.WriteString(".ReadUint64(")
	buf.WriteString(dst)
	buf.WriteString("[:])\n")
}

func (r *Int128Repr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".WriteUint128(")
	buf.WriteString(src)
	buf.WriteString("[:])\n")
}

func (r *Int128Repr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *Int128Repr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *Int128Repr) GoType() string {
	return "[16]byte"
}
func (r *Int128Repr) GoImports() []string {
	return nil
}

type Int256Repr struct {
}

func (r *Int256Repr) Resolve(resolver Resolver) {
}

func (r *Int256Repr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(reader)
	buf.WriteString(".ReadFull(")
	buf.WriteString(dst)
	buf.WriteString("[:])\n")
}

func (r *Int256Repr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(writer)
	buf.WriteString(".Write(")
	buf.WriteString(src)
	buf.WriteString("[:])\n")
}

func (r *Int256Repr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}

func (r *Int256Repr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *Int256Repr) GoType() string {
	return "[32]byte"
}
func (r *Int256Repr) GoImports() []string {
	return nil
}

type ObjectRepr struct {
}

func (r *ObjectRepr) Resolve(resolver Resolver) {
}
func (r *ObjectRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ReadFrom(")
	buf.WriteString(reader)
	buf.WriteString(")\n")
}
func (r *ObjectRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	// buf.WriteString(indent)
	// buf.WriteString("if ")
	// buf.WriteString(src)
	// buf.WriteString(" != nil {\b")

	buf.WriteString(indent)
	buf.WriteString(src)
	buf.WriteString(".WriteTo(")
	buf.WriteString(writer)
	buf.WriteString(")\n")

	// buf.WriteString(indent)
	// buf.WriteString("}\n")
}
func (r *ObjectRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
}
func (r *ObjectRepr) AppendGoDefs(buf *bytes.Buffer) {
}
func (r *ObjectRepr) GoType() string {
	return "tl.Struct"
}
func (r *ObjectRepr) GoImports() []string {
	return nil
}

type StructRepr struct {
	TLName string
	GoName string

	Ctor     *tlschema.Comb
	ArgReprs []ArgRepr
}

type ArgRepr struct {
	// Arg    *tlschema.Arg
	TLName     string
	GoName     string
	TypeRepr   Repr
	TLTypeName string
}

func (r *StructRepr) Resolve(resolver Resolver) {
	for _, arg := range r.Ctor.Args {
		ar := ArgRepr{
			// Arg:    &arg,
			TLName:     arg.Name,
			GoName:     tlschema.ToGoName(arg.Name),
			TypeRepr:   resolver.ResolveTypeExpr(arg.Type, r.TLName+":"+arg.Name),
			TLTypeName: arg.Type.String(),
		}
		r.ArgReprs = append(r.ArgReprs, ar)
	}
}

func (r *StructRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst, reader string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(".ReadFrom(")
	buf.WriteString(reader)
	buf.WriteString(")")
	buf.WriteString("\n")
}

func (r *StructRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src, writer string) {
	buf.WriteString(indent)
	buf.WriteString(src)
	buf.WriteString(".WriteTo(")
	buf.WriteString(writer)
	buf.WriteString(")\n")
}

func (r *StructRepr) AppendSwitchCase(buf *bytes.Buffer, indent, reader string) {
	buf.WriteString(indent)
	buf.WriteString("case ")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(":\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("s := new(")
	buf.WriteString(r.GoName)
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("s.ReadFrom(")
	buf.WriteString(reader)
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("return s\n")
}

func (r *StructRepr) AppendGoDefs(buf *bytes.Buffer) {
	buf.WriteString("\n")
	buf.WriteString("// ")
	buf.WriteString(r.GoName)
	buf.WriteString(" represents ")
	buf.WriteString(r.TLName)
	buf.WriteString(" from TL schema")
	buf.WriteString("\n")

	buf.WriteString("type ")
	buf.WriteString(r.GoName)
	buf.WriteString(" struct {\n")

	for _, ar := range r.ArgReprs {
		buf.WriteString("\t")
		buf.WriteString(ar.GoName)
		buf.WriteString(" ")
		buf.WriteString(ar.TypeRepr.GoType())
		buf.WriteString("  // ")
		buf.WriteString(ar.TLName)
		buf.WriteString(": ")
		buf.WriteString(ar.TLTypeName)
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") Cmd() uint32 {\n")
	buf.WriteString("\treturn ")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(";\n")
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") ReadFrom(r *tlschema.Reader) {\n")
	for _, ar := range r.ArgReprs {
		ar.TypeRepr.AppendReadStmt(buf, "\t", "s."+ar.GoName, "r")
	}
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") WriteTo(w *tlschema.Writer) {\n")
	buf.WriteString("\tw.WriteCmd(")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(")\n")
	for _, ar := range r.ArgReprs {
		ar.TypeRepr.AppendWriteStmt(buf, "\t", "s."+ar.GoName, "w")
	}
	buf.WriteString("}\n")
}

func (r *StructRepr) GoType() string {
	return "*" + r.GoName
}
func (r *StructRepr) GoImports() []string {
	var result []string
	for _, ar := range r.ArgReprs {
		result = append(result, ar.TypeRepr.GoImports()...)
	}
	return result
}

// func (r *StructRepr) GoDef(buf *bytes.Buffer) {
// 	// return "*" + r.GoName
// }
